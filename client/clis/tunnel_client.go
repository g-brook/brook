package clis

import (
	"errors"
	"fmt"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/google/uuid"
	"github.com/xtaci/smux"
	"io"
	"sync/atomic"
	"time"
)

type TunnelClientControl struct {
	Readers chan *exchange.Protocol

	Writers chan *exchange.Protocol

	Die chan struct{}

	RevStop chan struct{}
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
}

// BaseTunnelClient provides a base implementation of the TunnelClient interface.
type BaseTunnelClient struct {
	stream *smux.Stream

	cfg *configs.ClientTunnelConfig

	tcc *TunnelClientControl

	once atomic.Bool

	DoOpen func(stream *smux.Stream) error
}

func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig) *BaseTunnelClient {
	return &BaseTunnelClient{
		cfg: cfg,
		tcc: &TunnelClientControl{
			Die:     make(chan struct{}, 10),
			RevStop: make(chan struct{}, 1),
		},
	}
}

func (b *BaseTunnelClient) GetName() string {
	return "BaseTunnelClient"
}

func (b *BaseTunnelClient) Open(session *smux.Session) error {
	b.once.Store(false)
	stream, err := session.OpenStream()
	if err != nil {
		log.Error("Open session fail %v", err)
		return err
	}
	b.stream = stream
	go b.rveLoop()
	if b.DoOpen != nil {
		return b.DoOpen(stream)
	}
	panic("BaseTunnelClient: doOpen not set")
}

func (b *BaseTunnelClient) GetReaderWriter() io.ReadWriteCloser {
	return b.stream
}

func (b *BaseTunnelClient) Close() {
	_ = b.stream.Close()
	b.stream = nil
	b.tcc.Die <- struct{}{}
}

func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		BindId:     uuid.New().String(),
		TunnelPort: b.cfg.RemotePort,
		ProxyId:    "proxy1",
	}
}

func (b *BaseTunnelClient) Register() error {
	defer b.StopRev()
	req := b.GetRegisterReq()
	p, err := SyncWrite(req, 10*time.Second, func(bytes []byte) error {
		_, err := b.stream.Write(bytes)
		return err
	})
	if err != nil {
		return err
	}
	if p.RspCode != exchange.RspSuccess {
		b.once.Store(true)
		return errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	b.StopRev()
	return nil
}

func (b *BaseTunnelClient) StopRev() {
	b.tcc.RevStop <- struct{}{}
}

func (b *BaseTunnelClient) rveLoop() {
	read := func() error {
		pr, err := exchange.Decoder(b.stream)
		if err != nil {
			if err == io.EOF {
				return err
			}
			log.Error("Decoder %s error: %v", b.GetName(), err)
		} else {
			Tracker.Complete(pr.ReqId, pr)
		}
		return nil
	}
	for {
		select {
		case <-b.tcc.RevStop:
			return
		case <-b.stream.GetDieCh():
			return
		case <-b.tcc.Die:

			return
		default:
			if !b.once.Load() {
				if err := read(); err != nil {
					log.Error("Read %s error: %v", b.GetName(), err)
					return
				}
				b.once.Store(true)
			}
		}
	}
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
