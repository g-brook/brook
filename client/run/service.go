package run

import (
	"context"
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"sync"
	"time"
)

type Service struct {
	clis.BaseClientHandler
	ctx       context.Context
	connState chan struct{}
	connOnce  sync.Once
	manager   *clis.Transport
}

func (receiver *Service) Connection(_ *clis.ClientControl) {
	receiver.connOnce.Do(func() {
		close(receiver.connState)
	})
}

func NewService() *Service {
	return &Service{
		ctx:       context.Background(),
		connState: make(chan struct{}),
	}
}

func (receiver *Service) Run(cfg *configs.ClientConfig) context.Context {
	//Connection to server.
	receiver.manager = clis.NewTransport(cfg)
	//init manager transport.
	clis.InitManagerTransport(receiver.manager)
	receiver.manager.Connection(
		clis.WithTimeout(3*time.Second),
		clis.WithKeepAlive(10*time.Second),
		clis.WithClientHandler(receiver),
		clis.WithPingTime(cfg.PingTime*time.Millisecond),
		clis.WithClientHandler(clis.ManagerTransport),
	)
	<-receiver.connState
	//Update cli status.
	_ = receiver.connectionTunnel(cfg)
	return receiver.background()
}

func (receiver *Service) connectionTunnel(cfg *configs.ClientConfig) error {
	if cfg.Tunnels == nil {
		log.Warn("Tunnels is empty, no tunnels will be opened")
		return nil
	}
	req := exchange.QueryTunnelReq{}
	p, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		return err
	}
	rsp, err := exchange.Parse[exchange.QueryTunnelResp](p.Data)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	newCfg := configs.ClientConfig{
		ServerPort: rsp.TunnelPort,
		ServerHost: cfg.ServerHost,
		PingTime:   cfg.PingTime,
		Tunnels:    cfg.Tunnels,
	}
	//Start tunnel connection.
	tunnelTransport := clis.NewTransport(&newCfg)
	tunnelTransport.Connection(
		clis.WithPingTime(newCfg.PingTime*time.Millisecond),
		clis.WithClientSmux(clis.NewSmuxClientOption()))
	return nil
}

func (receiver *Service) background() context.Context {
	return receiver.ctx
}

type BrookClientHandler struct {
	clis.BaseClientHandler
}
