package tunnel

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
	"github.com/brook/server/srv"
)

type BaseTunnelServer struct {
	port    int
	Cfg     *configs.ServerTunnelConfig
	Server  *srv.Server
	DoStart func() error
}

// NewBaseTunnelServer 创建一个新的基础隧道服务器实例
func NewBaseTunnelServer(cfg *configs.ServerTunnelConfig) *BaseTunnelServer {
	return &BaseTunnelServer{
		port: cfg.Port,
		Cfg:  cfg,
	}
}

func (b *BaseTunnelServer) Start() error {
	b.Server = srv.NewServer(b.port)
	err := b.Server.Start()
	if err != nil {
		return err
	}
	if b.DoStart != nil {
		return b.DoStart()
	}
	panic("BaseTunnelServer not overite")
}

func (b *BaseTunnelServer) Port() int {
	return b.port
}

func (b *BaseTunnelServer) RegisterConn(v2 *transport.SChannel, request exchange.RegisterReqAndRsp) {

}

func (b *BaseTunnelServer) Receiver(v2 *transport.SChannel) {

}
