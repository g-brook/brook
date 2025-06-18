package run

import (
	"context"
	"github.com/brook/client/clients"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
	"time"
)

func init() {
	srv.RegisterTunnelClient("http", func() srv.TunnelClient {
		return &clients.HttpTunnelClient{}
	})
}

type Service struct {
	ctx   context.Context
	bgCtl chan struct{}
}

func NewService() *Service {
	return &Service{ctx: context.Background(),
		bgCtl: make(chan struct{}),
	}
}

func (receiver *Service) Run(cfg *configs.ClientConfig) error {
	//Connection to server.
	transport := srv.NewTransport(cfg)
	transport.Connection(
		srv.WithPingTime(cfg.PingTime * time.Millisecond))

	req := exchange.QueryTunnelReq{}
	p, _ := transport.WriteAsync(req, 5*time.Second)

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
	//Start tunnel connection.
	tunnelTransport := srv.NewTransport(&newCfg)
	tunnelTransport.Connection(
		srv.WithPingTime(newCfg.PingTime*time.Millisecond),
		srv.WithClientSmux(srv.NewSmuxClientOption()))

	return receiver.background()
}

func (receiver *Service) background() error {
	<-receiver.bgCtl
	return nil
}

type BrookClientHandler struct {
	srv.BaseClientHandler
}
