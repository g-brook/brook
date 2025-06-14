package run

import (
	"context"
	"github.com/brook/common/configs"
	"github.com/brook/common/srv"
	"time"
)

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
	transport := srv.NewTransport(1, cfg)
	transport.Connection(srv.WithPingTime(cfg.PingTime * time.Millisecond))
	//transport.Connection(srv.WithSmux(srv.NewSmuxClientOption()))
	return receiver.background()
}

func (receiver *Service) background() error {
	<-receiver.bgCtl
	return nil
}
