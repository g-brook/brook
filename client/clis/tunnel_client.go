package clis

import (
	"context"
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
	"io"
	"time"
)

type TunnelClientControl struct {
	cancelCtx context.Context
	cancel    context.CancelFunc
	Bucket    *exchange.MessageBucket
}

func NewTunnelClientControl(ctx context.Context) *TunnelClientControl {
	cancelCtx, cancel := context.WithCancel(ctx)
	return &TunnelClientControl{
		cancelCtx: cancelCtx,
		cancel:    cancel,
	}
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
				log.Warn("Open tunnel error...")
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

	// Open opens a tunnel using the provided session.
	// Parameters:
	//   - session: The smux session to use.
	// Returns:
	//   - error: An error if the tunnel could not be opened.
	Open(session *smux.Session) error

	// Close closes the tunnel.
	Close()

	// Done is done.
}

// BaseTunnelClient provides a base implementation of the TunnelClient interface.
type BaseTunnelClient struct {
	cfg *configs.ClientTunnelConfig

	tcc *TunnelClientControl

	DoOpen func(stream *smux.Stream) error
}

func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig) *BaseTunnelClient {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	return &BaseTunnelClient{
		cfg: cfg,
		tcc: &TunnelClientControl{
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

// Open is open
func (b *BaseTunnelClient) Open(session *smux.Session) error {
	openFunction := func() error {
		stream, err := session.OpenStream()
		if err != nil {
			log.Error("Open session fail %v", err)
			return err
		}
		bucket := exchange.NewMessageBucket(stream, b.tcc.cancelCtx)
		b.tcc.Bucket = bucket
		b.tcc.Bucket.Run()
		if b.DoOpen == nil {
			panic("DoOpen is nil")
		}
		err = b.DoOpen(stream)
		if err != nil {
			return err
		}
		<-bucket.Done()
		return nil
	}
	b.tcc.retry(openFunction)
	return nil
}

func (b *BaseTunnelClient) AddRead(cmd exchange.Cmd, read exchange.BucketRead) {
	b.tcc.Bucket.AddHandler(cmd, read)
}

func (b *BaseTunnelClient) Close() {
	b.tcc.cancel()
}

func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		BindId:     uuid.New().String(),
		TunnelPort: b.cfg.RemotePort,
		ProxyId:    "proxy1",
	}
}

func (b *BaseTunnelClient) Register() error {
	b.tcc.Bucket.AddHandler(exchange.Register, func(p *exchange.Protocol, _ io.ReadWriteCloser) {
		Tracker.Complete(p)
	},
	)
	req := b.GetRegisterReq()
	p, err := SyncWrite(req, 10*time.Second, func(p *exchange.Protocol) error {
		return b.tcc.Bucket.Push(p)
	})
	if err != nil {
		return err
	}
	if p.RspCode != exchange.RspSuccess {
		return errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	return nil
}

// TunnelsClient at all tunnel tunnel by map.
var TunnelsClient = make(map[string]FactoryFun)

// FactoryFun New tunnel client.
type FactoryFun func(config *configs.ClientTunnelConfig) TunnelClient

// RegisterTunnelClient
//
//	@Description: Register tunnel client.
//	@param name
//	@param factory
func RegisterTunnelClient(name string, factory FactoryFun) {
	TunnelsClient[name] = factory
}

// GetTunnelClient
//
//	@Description: Get tunnel client.
//	@param name
//	@return TunnelClient
func GetTunnelClient(name string, config *configs.ClientTunnelConfig) TunnelClient {
	fun := TunnelsClient[name]
	if fun != nil {
		return fun(config)
	}
	return nil
}

func GetTunnelClients() map[string]FactoryFun {
	return TunnelsClient
}
