package run

import (
	"context"
	"github.com/brook/common/configs"
	"github.com/brook/common/remote"
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
	transport := remote.NewTransport(1, cfg)
	transport.Connection(remote.WithSmux(remote.NewSmuxClientOption()))
	return receiver.background()
}

func (receiver *Service) background() error {
	<-receiver.bgCtl
	return nil
}
