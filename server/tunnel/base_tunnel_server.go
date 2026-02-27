/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tunnel

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/metrics"
	"github.com/g-brook/brook/server/srv"
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
	TunnelChannel   *hash.SyncMap[string, transport.Channel]
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
	return b.Cfg.Id
}

func (b *BaseTunnelServer) Clients() int {
	return b.TunnelChannel.Len()
}

func (b *BaseTunnelServer) TrafficObj() *metrics.TunnelTraffic {
	return b.trafficMetrics
}

func (b *BaseTunnelServer) ClientsInfo() []transport.Channel {
	return b.TunnelChannel.Values()
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
		b.TunnelChannel.Range(func(key string, value transport.Channel) (shouldContinue bool) {
			_ = value.Close()
			return false
		})
		b.TunnelChannel.Clear()
	}
	metrics.M.RemoveServer(b)
}

// NewBaseTunnelServer Create a new instance of the underlying tunnel server
func NewBaseTunnelServer(cfg *configs.ServerTunnelConfig) *BaseTunnelServer {
	return &BaseTunnelServer{
		port:           cfg.Port,
		Cfg:            cfg,
		TunnelChannel:  hash.NewSyncMap[string, transport.Channel](),
		ManagerChannel: hash.NewSyncSet[transport.Channel](),
		openCh:         make(chan error),
		handlers:       make(map[EventType]Event, 16),
		closeCtx:       context.Background(),
	}
}

func (b *BaseTunnelServer) Boot(_ srv.BootServer, _ srv.TraverseBy) error {
	b.openChOnce.Do(func() {
		close(b.openCh)
	})
	return nil
}

// Start  the tunnel server
func (b *BaseTunnelServer) Start(network lang.Network) error {
	threading.GoSafe(func() {
		b.Server = srv.NewServer(b.port)
		b.Server.AddHandler(b)
		err := b.Server.Start(srv.WithNetwork(network), srv.WithNewChannelFunc(func(ch transport.Channel) transport.Channel {
			return metrics.NewMetricsChannel(ch, b.trafficMetrics)
		}))
		if err != nil {
			log.Error("Start tunnel server port: error, %v:%v", err, b.Port())
			b.openCh <- err
		}
	})
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
	_ exchange.TRegister) (string, error) {
	oldCh, ok := b.TunnelChannel.Load(ch.GetId())
	if !ok || oldCh != ch {
		b.TunnelChannel.Store(ch.GetId(), ch)
		handler := b.handlers[Register]
		if handler != nil {
			handler(ch)
		}
		ch.OnClose(b.unRegister)
	}
	return ch.GetId(), nil
}

func (b *BaseTunnelServer) unRegister(ch transport.Channel) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.TunnelChannel.Delete(ch.GetId())
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

func (b *BaseTunnelServer) OpenWorker(transport.Channel, *exchange.ClientWorkConnReq) error {
	// Open a new goroutine to handle the channel
	return nil
}

func (b *BaseTunnelServer) Done() <-chan struct{} {
	return b.closeCtx.Done()
}

func (b *BaseTunnelServer) GetManager() transport.Channel {
	if b.ManagerChannel.Len() == 0 {
		log.Error("1:Tunnel Manager channel is empty. No available manager channel.")
		return nil
	}
	activeChannels := make([]transport.Channel, 0)
	b.ManagerChannel.ForEach(func(ch transport.Channel) bool {
		if !ch.IsClose() {
			activeChannels = append(activeChannels, ch)
		} else {
			b.ManagerChannel.Remove(ch)
		}
		return true
	})

	aLen := len(activeChannels)
	if aLen == 0 {
		log.Error("2:Tunnel Manager channel is empty. No available manager channel.")
		return nil
	}
	if aLen == 1 {
		return activeChannels[0]
	}
	sort.Slice(activeChannels, func(i, j int) bool {
		return activeChannels[i].ActiveTime().Before(activeChannels[j].ActiveTime())
	})

	v := index.Add(1) % uint64(len(activeChannels))
	return activeChannels[v]
}
