package tunnel

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/server/srv"
	"sync"
)

type BaseTunnelServer struct {
	srv.BaseServerHandler
	port       int
	Cfg        *configs.ServerTunnelConfig
	Server     *srv.Server
	DoStart    func() error
	Managers   map[string]transport.Channel
	openCh     chan error
	openChOnce sync.Once
}

// Shutdown  the tunnel server
func (b *BaseTunnelServer) Shutdown() {
	if b.Server != nil {
		b.Server.Shutdown()
	}
	if b.Managers != nil {
		clear(b.Managers)
		b.Managers = nil
	}
}

// NewBaseTunnelServer Create a new instance of the underlying tunnel server
func NewBaseTunnelServer(cfg *configs.ServerTunnelConfig) *BaseTunnelServer {
	return &BaseTunnelServer{
		port:     cfg.Port,
		Cfg:      cfg,
		Managers: make(map[string]transport.Channel),
		openCh:   make(chan error),
	}
}
func (b *BaseTunnelServer) Boot(_ *srv.Server, _ srv.TraverseBy) {
	b.openChOnce.Do(func() {
		close(b.openCh)
	})
}

// Start  the tunnel server
func (b *BaseTunnelServer) Start() error {
	go func() {
		b.Server = srv.NewServer(b.port)
		b.Server.AddHandler(b)
		err := b.Server.Start()
		if err != nil {
			log.Error("Start tunnel server port: error,", b.Port())
			b.openCh <- err
		}
	}()
	if err := <-b.openCh; err != nil {
		return err
	}
	if b.DoStart != nil {
		return b.DoStart()
	}
	panic("BaseTunnelServer not overite")
}

// Port  the tunnel server port
func (b *BaseTunnelServer) Port() int {
	return b.port
}

// RegisterConn  register the tunnel server connection
func (b *BaseTunnelServer) RegisterConn(ch transport.Channel, request exchange.RegisterReqAndRsp) {
	b.Managers[request.BindId] = ch
}
