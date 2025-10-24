package tunnel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/brook/server/metrics"
	"github.com/brook/server/srv"
)

type EventType int

var (
	Unregister EventType = 1

	Register EventType = 2

	index atomic.Uint64
)

type Event func(ch transport.Channel)

type UpdateConfigFunction func(cfg *configs.ServerTunnelConfig)

type BaseTunnelServer struct {
	srv.BaseServerHandler
	port            int
	Cfg             *configs.ServerTunnelConfig
	Server          *srv.Server
	DoStart         func() error
	TunnelChannel   map[string]transport.Channel
	ManagerChannel  *hash.SyncSet[transport.Channel]
	openCh          chan error
	openChOnce      sync.Once
	handlers        map[EventType]Event
	lock            sync.Mutex
	closeCtx        context.Context
	trafficMetrics  *metrics.TunnelTraffic
	runtime         time.Time
	UpdateConfigFun UpdateConfigFunction
}

func (b *BaseTunnelServer) Id() string {
	return b.Cfg.Id

}
func (b *BaseTunnelServer) Type() string {
	return string(b.Cfg.Type)
}

func (b *BaseTunnelServer) Connections() int {
	return len(b.Server.Connections())
}

func (b *BaseTunnelServer) Name() string {
	return "tunnel"
}

func (b *BaseTunnelServer) Users() int {
	return len(b.TunnelChannel)
}

func (b *BaseTunnelServer) TrafficObj() *metrics.TunnelTraffic {
	return b.trafficMetrics
}

func (b *BaseTunnelServer) AddEvent(etype EventType,
	event Event) {
	b.handlers[etype] = event
}

func (b *BaseTunnelServer) PutManager(ch transport.Channel) {
	b.ManagerChannel.Add(ch)
	ch.OnClose(func(channel transport.Channel) {
		b.ManagerChannel.Remove(channel)
	})
}

// Shutdown  the tunnel server
func (b *BaseTunnelServer) Shutdown() {
	if b.Server != nil {
		b.Server.Shutdown(b.closeCtx)
	}
	if b.TunnelChannel != nil {
		for t := range b.TunnelChannel {
			_ = b.TunnelChannel[t].Close()
		}
		clear(b.TunnelChannel)
	}
	metrics.M.RemoveServer(b)
}

// NewBaseTunnelServer Create a new instance of the underlying tunnel server
func NewBaseTunnelServer(cfg *configs.ServerTunnelConfig) *BaseTunnelServer {
	return &BaseTunnelServer{
		port:           cfg.Port,
		Cfg:            cfg,
		TunnelChannel:  make(map[string]transport.Channel),
		ManagerChannel: hash.NewSyncSet[transport.Channel](),
		openCh:         make(chan error),
		handlers:       make(map[EventType]Event, 16),
		closeCtx:       context.Background(),
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
		err := b.Server.Start(srv.WithNetwork(network), srv.WithNewChannelFunc(func(ch *srv.GChannel) transport.Channel {
			return metrics.NewMetricsChannel(ch, b.trafficMetrics)
		}))
		if err != nil {
			log.Error("Start tunnel server port: error, %v:%v", err, b.Port())
			b.openCh <- err
		}
	}()
	if err := <-b.openCh; err != nil {
		return err
	}
	if b.DoStart != nil {
		b.runtime = time.Now()
		b.trafficMetrics = metrics.M.PutServer(b)
		return b.DoStart()
	}
	return nil
}

// Port  the tunnel server port
func (b *BaseTunnelServer) Port() int {
	return b.port
}

func (b *BaseTunnelServer) Runtime() time.Time {
	return b.runtime
}

// RegisterConn  register the tunnel server connection
func (b *BaseTunnelServer) RegisterConn(ch transport.Channel,
	_ exchange.TRegister) {
	oldCh, ok := b.TunnelChannel[ch.GetId()]
	if !ok || oldCh != ch {
		b.TunnelChannel[ch.GetId()] = ch
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
	delete(b.TunnelChannel, ch.GetId())
	handler := b.handlers[Unregister]
	if handler != nil {
		handler(ch)
	}
}
func (b *BaseTunnelServer) UpdateConfig(config *configs.ServerTunnelConfig) {
	if b.UpdateConfigFun != nil {
		b.UpdateConfigFun(config)
	}
}

func (b *BaseTunnelServer) Done() <-chan struct{} {
	return b.closeCtx.Done()
}

func (b *BaseTunnelServer) GetManager() transport.Channel {
	if b.ManagerChannel.Len() == 0 {
		return nil
	}
	i := b.ManagerChannel.Len()
	v := index.Add(1) % uint64(i)
	bt := 0
br:
	channel := b.ManagerChannel.List()[v]
	if channel.IsClose() && bt < i {
		b.ManagerChannel.Remove(channel)
		bt++
		goto br
	}
	return channel
}
