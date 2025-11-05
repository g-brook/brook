/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
	"github.com/xtaci/smux"
)

type TunnelClientControl struct {
	cancelCtx context.Context
	cancel    context.CancelFunc
	Bucket    *exchange.MessageBucket
}

var (
	sessionError = errors.New("session error")
)

// Context returns the context associated with the TunnelClientControl.
// This context can be used to track the lifecycle of the tunnel client
// and to cancel operations if needed.
//
// Parameters:
//
//	receiver - A pointer to the TunnelClientControl instance on which the method is called
//
// Returns:
//
//	context.Context - The context associated with the TunnelClientControl
func (receiver *TunnelClientControl) Context() context.Context {
	return receiver.cancelCtx
}

// Cancel cancels the tunnel client operation by calling the underlying cancel function.
// It takes a receiver of type TunnelClientControl as a pointer to modify the instance.
func (receiver *TunnelClientControl) Cancel() {
	// Call the cancel function associated with the TunnelClientControl instance
	receiver.cancel()
}

// retry is a method that takes a function f of type func() error as a parameter
// It is designed to repeatedly execute the function f with a fixed interval
// until the context associated with the TunnelClientControl is cancelled
func (receiver *TunnelClientControl) retry(f func() error) {
	threading.GoSafe(func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-receiver.Context().Done():
				return
			default:
			}
			if err := f(); err != nil {
				log.Warn("Active tunnel error...", err)
				if errors.Is(err, sessionError) {
					return
				}
			}
			ticker.Reset(time.Second * 5)
			select {
			case <-receiver.Context().Done():
				return
			case <-ticker.C:
			}
		}
	})
}

// TunnelClient defines the interface for a tunnel client.
type TunnelClient interface {
	// GetName returns the name of the tunnel client.
	// Returns:
	//   - string: The name of the tunnel client.
	GetName() string

	// Open Active opens a tunnel using the provided session.
	// Parameters:
	//   - session: The smux session to use.
	// Returns:
	//   - error: An error if the tunnel could not be opened.
	Open(session *smux.Session) error

	Done() <-chan struct{}

	// Close closes the tunnel.
	Close()
}

// BaseTunnelClient provides a base implementation of the TunnelClient interface.
type BaseTunnelClient struct {
	cfg *configs.ClientTunnelConfig

	TcControl *TunnelClientControl

	DoOpen func(stream *transport.SChannel) error

	DoRelease func(stream *transport.SChannel) error

	session *smux.Session

	isRetryOpen bool

	isOpen bool
}

// NewBaseTunnelClient creates a new BaseTunnelClient instance with the provided configuration
// and determines if it should be automatically opened.
//
// Parameters:
//   - cfg: A pointer to ClientTunnelConfig containing the tunnel configuration
//   - isAutoOpen: A boolean flag indicating whether the tunnel should be opened automatically
//
// Returns:
//   - A pointer to the newly created BaseTunnelClient instance
func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig, isRetryOpen bool) *BaseTunnelClient {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	return &BaseTunnelClient{
		cfg:         cfg,
		isRetryOpen: isRetryOpen,
		TcControl: &TunnelClientControl{
			cancelCtx: cancelCtx,
			cancel:    cancelFunc,
		},
	}
}

// GetCfg is get cfg
// GetCfg retrieves the client tunnel configuration from the BaseTunnelClient
// This method provides read-only access to the tunnel configuration
func (b *BaseTunnelClient) GetCfg() *configs.ClientTunnelConfig {
	return b.cfg // Return the stored configuration
}

// GetName is get name
// GetName is a method of BaseTunnelClient struct that returns the name of the client
// It takes no parameters and returns a string value
func (b *BaseTunnelClient) GetName() string {
	return "BaseTunnelClient" // Return the string "BaseTunnelClient" as the client name
}

func (b *BaseTunnelClient) Open(session *smux.Session) error {
	b.session = session
	return b.OpenStream()
}

// OpenStream opens a new stream for the BaseTunnelClient
// It handles both automatic and manual stream opening based on the IsAutoOpen flag
func (b *BaseTunnelClient) OpenStream() error {
	if b.isOpen {
		return fmt.Errorf("tunnel is open")
	}
	openFunction := func() error {
		stream, err := b.session.OpenStream()
		if b.session.IsClosed() {
			_ = b.session.Close()
			log.Debug("session is close, exit")
			b.TcControl.Cancel()
			return sessionError
		}
		if err != nil {
			log.Error("Active session fail %v", err)
			return sessionError
		}
		streamCancelCtx, streamCancel := context.WithCancel(b.TcControl.Context())
		channel := transport.NewSChannel(stream, streamCancelCtx, true)
		bucket := exchange.NewMessageBucket(channel, channel.Ctx())
		b.TcControl.Bucket = bucket
		bucket.Run()
		err = b.DoOpen(channel)
		if err != nil {
			bucket.Close()
			streamCancel()
			return err
		}
		b.isOpen = true
		<-channel.Done()
		log.Info("Tunnel stream close exit:%v:%v", stream.RemoteAddr(), stream.ID())
		streamCancel()
		if !b.isRetryOpen {
			b.release(channel)
		}
		return nil
	}
	// If is retry open is enabled, use retry mechanism
	if b.isRetryOpen {
		b.TcControl.retry(openFunction)
	} else {
		return openFunction()
	}

	return nil
}

func (b *BaseTunnelClient) release(ch *transport.SChannel) {
	if b.DoRelease != nil {
		_ = b.DoRelease(ch)
	}
	_ = ch.Close()
	b.TcControl.Cancel()
	b.TcControl.Bucket.Close()
	b.TcControl = nil
	b.isOpen = false
}

// AddReadHandler adds a read handler for a specific command to the tunnel client
// Parameters:
//   - cmd: The command type to handle
//   - read: The read handler function that processes the command
func (b *BaseTunnelClient) AddReadHandler(cmd exchange.Cmd, read exchange.BucketRead) {
	b.TcControl.Bucket.AddHandler(cmd, read)
}

func (b *BaseTunnelClient) Done() <-chan struct{} {
	return b.TcControl.Context().Done()
}

func (b *BaseTunnelClient) Close() {
	b.TcControl.cancel()
}

// GetRegisterReq returns a RegisterReqAndRsp struct with configuration data from the BaseTunnelClient
// This method is used to prepare registration request parameters for the tunnel connection
func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		TunnelPort: b.GetCfg().RemotePort, // Set the tunnel port from configuration
		ProxyId:    b.GetCfg().ProxyId,    // Set the proxy identifier from configuration
		TunnelType: b.GetCfg().TunnelType, // Set the tunnel type from configuration
		HttpId:     b.GetCfg().HttpId,
	}
}

// Register is a method of BaseTunnelClient that handles the registration process
// It sends a registration request and processes the response
// Returns the registration result or an error if any step fails
func (b *BaseTunnelClient) Register(req exchange.InBound) (*exchange.RegisterReqAndRsp, error) {
	if req == nil {
		req = b.GetRegisterReq()
	}
	p, err := b.TcControl.Bucket.SyncPushWithRequest(req)
	if err != nil {
		// Return error if the request fails
		return nil, err
	}
	if p.RspCode != exchange.RspSuccess {
		// Return error if the response code is not success
		return nil, errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	result, err := exchange.Parse[exchange.RegisterReqAndRsp](p.Data)
	if err != nil {
		// Return error if parsing fails
		return nil, err
	}
	return result, nil
}

// AsyncRegister is an asynchronous method that registers a callback handler for incoming messages
// and sends a registration request to the server
//
// Parameters:
//
//	readCallBack: A callback function of type exchange.BucketRead that will be invoked
//	             when messages are received for this client
//
// Returns:
//
//	error: Any error that occurred during the registration process, or nil if successful
func (b *BaseTunnelClient) AsyncRegister(req exchange.InBound, readCallBack exchange.BucketRead) error {
	if req == nil {
		req = b.GetRegisterReq()
	}
	b.TcControl.Bucket.AddHandler(req.Cmd(), readCallBack)
	return b.TcControl.Bucket.PushWitchRequest(req)
}

// TunnelsClient at all tunnel tunnel by map.
var TunnelsClient = make(map[lang.TunnelType]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func(config *configs.ClientTunnelConfig) TunnelClient

// RegisterTunnelClient registers a tunnel client factory for a specific tunnel type
// This function allows the system to create and manage different types of tunnel clients
//
// Parameters:
//
//	name - the type of tunnel client to register (utils.TunnelType)
//	factory - the factory function that creates instances of the tunnel client (FactoryFun)
func RegisterTunnelClient(name lang.TunnelType, factory FactoryFun) {
	TunnelsClient[name] = factory // Store the factory function in the map with the tunnel type as the key
}

// GetTunnelClient creates and returns a tunnel client based on the provided tunnel type and configuration
// Parameters:
//   - name: The type of tunnel to create (utils.TunnelType)
//   - configs: Configuration for the tunnel client (configs.ClientTunnelConfig)
//
// Returns:
//   - TunnelClient: The created tunnel client instance, or nil if the tunnel type is not supported
func GetTunnelClient(name lang.TunnelType, config *configs.ClientTunnelConfig) TunnelClient {
	// Retrieve the tunnel client constructor function from the TunnelsClient map
	fun := TunnelsClient[name]
	// If a constructor function exists for the specified tunnel type
	if fun != nil {
		// Create and return the tunnel client using the provided configuration
		return fun(config)
	}
	// Return nil if the tunnel type is not supported
	return nil
}

// GetTunnelClients returns a map of tunnel types to their corresponding factory functions
// This function provides access to the available tunnel client implementations
func GetTunnelClients() map[lang.TunnelType]FactoryFun {
	return TunnelsClient // Return the global map of tunnel clients
}
