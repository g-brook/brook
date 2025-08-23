package clis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
)

type TunnelClientControl struct {
	cancelCtx context.Context
	cancel    context.CancelFunc
	Bucket    *exchange.MessageBucket
}

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
	// Launch a new goroutine to handle the retry mechanism
	go func() {
		// Create a ticker that will trigger every 5 seconds
		ticker := time.NewTicker(time.Second * 5)
		// Ensure the ticker is stopped when this function exits
		defer ticker.Stop()
		// Infinite loop to continuously attempt function execution
		for {
			// Check if the context has been cancelled
			select {
			case <-receiver.Context().Done():
				return
			default:
				// Continue with execution if context is not cancelled
			}
			// Execute the provided function and handle any errors
			if err := f(); err != nil {
				log.Warn("Active tunnel error...")
			}
			// Reset the ticker to maintain consistent 5-second intervals
			ticker.Reset(time.Second * 5)
			// Wait for either context cancellation or the next ticker event
			select {
			case <-receiver.Context().Done():
				return
			case <-ticker.C:
				// Continue to next iteration when ticker fires
			}
		}
	}()
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

	// Close closes the tunnel.
	Close()
}

// BaseTunnelClient provides a base implementation of the TunnelClient interface.
type BaseTunnelClient struct {
	cfg *configs.ClientTunnelConfig

	Tcc *TunnelClientControl

	DoOpen func(stream *transport.SChannel) error

	IsAutoOpen bool

	session *smux.Session
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
func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig, isAutoOpen bool) *BaseTunnelClient {
	// Create a cancellable context and its corresponding cancel function
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	// Initialize and return a new BaseTunnelClient with the provided configuration
	// and a new TunnelClientControl with the cancellable context
	return &BaseTunnelClient{
		cfg:        cfg,        // Assign the provided configuration
		IsAutoOpen: isAutoOpen, // Set the auto-open flag
		Tcc: &TunnelClientControl{ // Initialize the tunnel client control structure
			cancelCtx: cancelCtx,  // Store the cancellable context
			cancel:    cancelFunc, // Store the cancel function
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

// Open Active is OpenStream
// Open establishes a connection using the provided session
// It sets the session for the BaseTunnelClient and optionally opens a stream
// if auto-open is enabled.
//
// Parameters:
//
//	session - A pointer to a smux.Session representing the multiplexed session
//
// Returns:
//
//	error - If auto-open is enabled and stream opening fails, returns the error;
//	        otherwise returns nil
func (b *BaseTunnelClient) Open(session *smux.Session) error {
	// Set the session for the BaseTunnelClient
	b.session = session
	// Check if auto-open is enabled
	if b.IsAutoOpen {
		// If enabled, attempt to open a stream and return any error
		return b.OpenStream()
	}
	// If auto-open is disabled, return nil
	return nil
}

// OpenStream opens a new stream for the BaseTunnelClient
// It handles both automatic and manual stream opening based on the IsAutoOpen flag
func (b *BaseTunnelClient) OpenStream() error {
	// Define an inner function that contains the actual stream opening logic
	openFunction := func() error {
		// Try to open a new stream from the session
		stream, err := b.session.OpenStream()
		// Check if the session is closed, return early if true
		if b.session.IsClosed() {
			return nil
		}
		// Handle any errors that occurred during stream opening
		if err != nil {
			log.Error("Active session fail %v", err)
			return err
		}
		// Create a new secure channel for the stream
		channel := transport.NewSChannel(stream, b.Tcc.cancelCtx, true)
		// Close any existing bucket and reset it to nil
		if b.Tcc.Bucket != nil {
			b.Tcc.Bucket.Close()
			b.Tcc.Bucket = nil
		}
		// Create a new message bucket for the channel
		bucket := exchange.NewMessageBucket(channel, b.Tcc.cancelCtx)
		b.Tcc.Bucket = bucket
		// Start the bucket processing
		b.Tcc.Bucket.Run()
		// Check if DoOpen is properly initialized
		if b.DoOpen == nil {
			panic("DoOpen is nil")
		}
		// Execute the DoOpen callback with the new channel
		err = b.DoOpen(channel)
		if err != nil {
			return err
		}
		// Wait for the bucket to complete processing
		<-bucket.Done()
		// Log the stream closure information
		log.Info("Tunnel stream close exit:%v:%v", stream.RemoteAddr(), stream.ID())
		return nil
	}
	// Handle automatic or manual stream opening based on IsAutoOpen flag
	if b.IsAutoOpen {
		// If auto-open is enabled, use retry mechanism
		b.Tcc.retry(openFunction)
		return nil
	} else {
		// Otherwise, directly execute and return the result of openFunction
		return openFunction()
	}
}

// AddReadHandler adds a read handler for a specific command to the tunnel client
// Parameters:
//   - cmd: The command type to handle
//   - read: The read handler function that processes the command
func (b *BaseTunnelClient) AddReadHandler(cmd exchange.Cmd, read exchange.BucketRead) {
	// Add the handler to the tunnel's command channel
	b.Tcc.Bucket.AddHandler(cmd, read)
}

func (b *BaseTunnelClient) Close() {
	b.Tcc.cancel()
}

// GetRegisterReq returns a RegisterReqAndRsp struct with configuration data from the BaseTunnelClient
// This method is used to prepare registration request parameters for the tunnel connection
func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		TunnelPort: b.GetCfg().RemotePort, // Set the tunnel port from configuration
		ProxyId:    b.GetCfg().ProxyId,    // Set the proxy identifier from configuration
		TunnelType: b.GetCfg().TunnelType, // Set the tunnel type from configuration
	}
}

// Register is a method of BaseTunnelClient that handles the registration process
// It sends a registration request and processes the response
// Returns the registration result or an error if any step fails
func (b *BaseTunnelClient) Register() (*exchange.RegisterReqAndRsp, error) {
	// Get the registration request from the client
	req := b.GetRegisterReq()
	// Sync push the request to the TCC bucket and get the response
	p, err := b.Tcc.Bucket.SyncPushWithRequest(req)
	if err != nil {
		// Return error if the request fails
		return nil, err
	}
	// Check if the response code indicates success
	if p.RspCode != exchange.RspSuccess {
		// Return error if the response code is not success
		return nil, errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	// Parse the response data into RegisterReqAndRsp struct
	result, err := exchange.Parse[exchange.RegisterReqAndRsp](p.Data)
	if err != nil {
		// Return error if parsing fails
		return nil, err
	}
	// Return the parsed result
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
func (b *BaseTunnelClient) AsyncRegister(readCallBack exchange.BucketRead) error {
	// Create a new registration request using the client's configuration
	req := b.GetRegisterReq()
	// Register the callback handler for the specific command type in the bucket
	b.Tcc.Bucket.AddHandler(req.Cmd(), readCallBack)
	// Push the registration request to the server
	return b.Tcc.Bucket.PushWitchRequest(req)
}

// TunnelsClient at all tunnel tunnel by map.
var TunnelsClient = make(map[utils.TunnelType]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func(config *configs.ClientTunnelConfig) TunnelClient

// RegisterTunnelClient registers a tunnel client factory for a specific tunnel type
// This function allows the system to create and manage different types of tunnel clients
//
// Parameters:
//
//	name - the type of tunnel client to register (utils.TunnelType)
//	factory - the factory function that creates instances of the tunnel client (FactoryFun)
func RegisterTunnelClient(name utils.TunnelType, factory FactoryFun) {
	TunnelsClient[name] = factory // Store the factory function in the map with the tunnel type as the key
}

// GetTunnelClient creates and returns a tunnel client based on the provided tunnel type and configuration
// Parameters:
//   - name: The type of tunnel to create (utils.TunnelType)
//   - config: Configuration for the tunnel client (configs.ClientTunnelConfig)
//
// Returns:
//   - TunnelClient: The created tunnel client instance, or nil if the tunnel type is not supported
func GetTunnelClient(name utils.TunnelType, config *configs.ClientTunnelConfig) TunnelClient {
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
func GetTunnelClients() map[utils.TunnelType]FactoryFun {
	return TunnelsClient // Return the global map of tunnel clients
}
