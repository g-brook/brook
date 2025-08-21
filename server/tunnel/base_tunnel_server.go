package tunnel

import (
	"context"
	"sync"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/brook/server/srv"
)

type EventType int

var (
	Unregister EventType = 1

	Register EventType = 2
)

type Event func(ch transport.Channel)

type BaseTunnelServer struct {
	srv.BaseServerHandler
	port       int
	Cfg        *configs.ServerTunnelConfig
	Server     *srv.Server
	DoStart    func() error
	Managers   map[string]transport.Channel
	openCh     chan error
	openChOnce sync.Once
	handlers   map[EventType]Event
	lock       sync.Mutex
	closeCtx   context.Context
}

func (b *BaseTunnelServer) AddEvent(etype EventType,
	event Event) {
	b.handlers[etype] = event
}

// Shutdown  the tunnel server
func (b *BaseTunnelServer) Shutdown() {
	if b.Server != nil {
		b.Server.Shutdown(b.closeCtx)
	}
	if b.Managers != nil {
		for t := range b.Managers {
			_ = b.Managers[t].Close()
		}
		clear(b.Managers)
	}
}

// NewBaseTunnelServer Create a new instance of the underlying tunnel server
func NewBaseTunnelServer(cfg *configs.ServerTunnelConfig) *BaseTunnelServer {
	return &BaseTunnelServer{
		port:     cfg.Port,
		Cfg:      cfg,
		Managers: make(map[string]transport.Channel),
		openCh:   make(chan error),
		handlers: make(map[EventType]Event, 16),
		closeCtx: context.Background(),
	}
}
func (b *BaseTunnelServer) Boot(_ *srv.Server, _ srv.TraverseBy) {
	b.openChOnce.Do(func() {
		close(b.openCh)
	})
}

// Start  the tunnel server
func (b *BaseTunnelServer) Start(network utils.Network) error {
	go func() {
		b.Server = srv.NewServer(b.port)
		b.Server.AddHandler(b)
		err := b.Server.Start(srv.WithNetwork(network))
		if err != nil {
			log.Error("Start tunnel server port: error, %v:%v", err, b.Port())
			b.openCh <- err
		}
	}()
	if err := <-b.openCh; err != nil {
		return err
	}
	if b.DoStart != nil {
		return b.DoStart()
	}
	return nil
}

// Port  the tunnel server port
func (b *BaseTunnelServer) Port() int {
	return b.port
}

// RegisterConn  register the tunnel server connection
func (b *BaseTunnelServer) RegisterConn(ch transport.Channel,
	_ exchange.RegisterReqAndRsp) {
	oldCh, ok := b.Managers[ch.GetId()]
	if !ok || oldCh != ch {
		b.Managers[ch.GetId()] = ch
		handler := b.handlers[Register]
		if handler != nil {
			handler(ch)
		}
		ch.OnClose(b.unRegister)
	}
}

func (b *BaseTunnelServer) unRegister(ch transport.Channel) {
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.Managers, ch.GetId())
	handler := b.handlers[Unregister]
	if handler != nil {
		handler(ch)
	}
}

func (b *BaseTunnelServer) Done() <-chan struct{} {
	return b.closeCtx.Done()
}
