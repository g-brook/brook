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
	"time"
)

type TunnelClientControl struct {
	Readers chan *exchange.Protocol

	Writers chan *exchange.Protocol

	Die chan struct{}

	RevStop chan struct{}
}

type TunnelClient interface {

	//
	// GetName
	//  @Description: Get name.
	//  @return string
	//
	GetName() string

	//
	// Open
	//  @Description: Open tunnel.
	//  @param session
	//
	Open(session *smux.Session) error

	//
	// Close
	//  @Description: Close
	//  @param session
	//
	Close()
}

// BaseTunnelClient is base impl.
type BaseTunnelClient struct {
	stream *smux.Stream

	cfg *configs.ClientTunnelConfig

	tcc *TunnelClientControl

	DoOpen func(stream *smux.Stream) error
}

func NewBaseTunnelClient(cfg *configs.ClientTunnelConfig) *BaseTunnelClient {
	return &BaseTunnelClient{
		cfg: cfg,
		tcc: &TunnelClientControl{
			Readers: make(chan *exchange.Protocol),
			Writers: make(chan *exchange.Protocol),
			RevStop: make(chan struct{}),
			Die:     make(chan struct{}),
		},
	}
}

func (b *BaseTunnelClient) GetName() string {
	return "BaseTunnelClient"
}

func (b *BaseTunnelClient) Open(session *smux.Session) error {
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

func (b *BaseTunnelClient) Close() {
	_ = b.stream.Close()
	b.tcc.RevStop <- struct{}{}
	b.tcc.Die <- struct{}{}
}

func (b *BaseTunnelClient) GetRegisterReq() exchange.RegisterReqAndRsp {
	return exchange.RegisterReqAndRsp{
		BindId:     uuid.New().String(),
		TunnelPort: b.cfg.RemotePort,
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
		return errors.New(fmt.Sprintf("register error: %d", p.RspCode))
	}
	return nil
}

func (b *BaseTunnelClient) StopRev() {
	b.tcc.RevStop <- struct{}{}
}

func (b *BaseTunnelClient) rveLoop() {

	for {
		select {
		case <-b.tcc.RevStop:
			return
		case <-b.stream.GetDieCh():
			return
		default:
		}
		pr, err := exchange.Decoder(b.stream)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Error("Decoder %s error: %v", b.GetName(), err)
		} else {
			Tracker.Complete(pr.ReqId, pr)
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
