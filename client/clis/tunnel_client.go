package clis

import (
	"context"
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/xtaci/smux"
	"time"
)

type TunnelClientControl struct {
	cancelCtx context.Context
	cancel    context.CancelFunc
	Bucket    *exchange.MessageBucket
}

func (receiver *TunnelClientControl) Context() context.Context {
	return receiver.cancelCtx
}

func (receiver *TunnelClientControl) Cancel() {
	receiver.cancel()
}

func (receiver *TunnelClientControl) retry(f func() error) {
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-receiver.Context().Done():
				return
			default:
			}
			if err := f(); err != nil {
				log.Warn("Active tunnel error...")
			}
			ticker.Reset(time.Second * 5)
			select {
			case <-receiver.Context().Done():
				return
			case <-ticker.C:
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
}

func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig) *BaseTunnelClient {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	return &BaseTunnelClient{
		cfg: cfg,
		Tcc: &TunnelClientControl{
			cancelCtx: cancelCtx,
			cancel:    cancelFunc,
		},
	}
}

// GetCfg is get cfg
func (b *BaseTunnelClient) GetCfg() *configs.ClientTunnelConfig {
	return b.cfg
}

// GetName is get name
func (b *BaseTunnelClient) GetName() string {
	return "BaseTunnelClient"
}

// Open Active is open
func (b *BaseTunnelClient) Open(session *smux.Session) error {
	openFunction := func() error {
		stream, err := session.OpenStream()
		if session.IsClosed() {
			return nil
		}
		if err != nil {
			log.Error("Active session fail %v", err)
			return err
		}
		if b.Tcc.Bucket != nil {
			b.Tcc.Bucket.Close()
			b.Tcc.Bucket = nil
		}
		channel := transport.NewSChannel(stream, b.Tcc.cancelCtx, true)
		bucket := exchange.NewMessageBucket(channel, b.Tcc.cancelCtx)
		b.Tcc.Bucket = bucket
		b.Tcc.Bucket.Run()
		if b.DoOpen == nil {
			panic("DoOpen is nil")
		}
		err = b.DoOpen(channel)
		if err != nil {
			return err
		}
		<-bucket.Done()
		return nil
	}
	b.Tcc.retry(openFunction)
	return nil
}

func (b *BaseTunnelClient) AddReadHandler(cmd exchange.Cmd, read exchange.BucketRead) {
	b.Tcc.Bucket.AddHandler(cmd, read)
}

func (b *BaseTunnelClient) Close() {
	b.Tcc.cancel()
}

func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		TunnelPort: b.GetCfg().RemotePort,
		ProxyId:    b.GetCfg().ProxyId,
		TunnelType: b.GetCfg().Type,
	}
}

func (b *BaseTunnelClient) Register() (*exchange.RegisterReqAndRsp, error) {
	req := b.GetRegisterReq()
	p, err := b.Tcc.Bucket.SyncPushWithRequest(req)
	if err != nil {
		return nil, err
	}
	if p.RspCode != exchange.RspSuccess {
		return nil, errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	result, err := exchange.Parse[exchange.RegisterReqAndRsp](p.Data)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TunnelsClient at all tunnel tunnel by map.
var TunnelsClient = make(map[utils.TunnelType]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func(config *configs.ClientTunnelConfig) TunnelClient

// RegisterTunnelClient
//
//	@Description: Register tunnel client.
//	@param name
//	@param factory
func RegisterTunnelClient(name utils.TunnelType, factory FactoryFun) {
	TunnelsClient[name] = factory
}

// GetTunnelClient
//
//	@Description: Get tunnel client.
//	@param name
//	@return TunnelClient
func GetTunnelClient(name utils.TunnelType, config *configs.ClientTunnelConfig) TunnelClient {
	fun := TunnelsClient[name]
	if fun != nil {
		return fun(config)
	}
	return nil
}

func GetTunnelClients() map[utils.TunnelType]FactoryFun {
	return TunnelsClient
}
