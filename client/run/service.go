package run

import (
	"context"
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"sync"
	"time"
)

type Service struct {
	clis.BaseClientHandler
	ctx       context.Context
	bgCtl     chan struct{}
	connState chan struct{}
	connOnce  sync.Once
}

func (receiver *Service) Connection(_ *clis.ClientControl) {
	receiver.connOnce.Do(func() {
		close(receiver.connState)
	})
}

func NewService() *Service {
	return &Service{ctx: context.Background(),
		bgCtl:     make(chan struct{}),
		connState: make(chan struct{}),
	}
}

func (receiver *Service) Run(cfg *configs.ClientConfig) error {
	//Connection to server.
	transport := clis.NewTransport(cfg)
	transport.Connection(
		clis.WithClientHandler(receiver),
		clis.WithPingTime(cfg.PingTime*time.Millisecond))
	<-receiver.connState
	_ = receiver.openTunnel(cfg, transport)
	return receiver.background()
}

func (receiver *Service) openTunnel(cfg *configs.ClientConfig, transport *clis.Transport) error {
	if cfg.Tunnels == nil {
		log.Warn("Tunnels is empty, no tunnels will be opened")
		return nil
	}
	req := exchange.QueryTunnelReq{}
	p, err := transport.WriteAsync(req, 5*time.Second)
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
	}
	newCfg.Tunnels = append(cfg.Tunnels, &configs.ClientTunnelConfig{
		Type: utils.EchoTest,
	})
	//Start tunnel connection.
	tunnelTransport := clis.NewTransport(&newCfg)
	tunnelTransport.Connection(
		clis.WithPingTime(newCfg.PingTime*time.Millisecond),
		clis.WithClientSmux(clis.NewSmuxClientOption()))
	return nil
}

func (receiver *Service) background() error {
	<-receiver.bgCtl
	return nil
}

type BrookClientHandler struct {
	clis.BaseClientHandler
}
