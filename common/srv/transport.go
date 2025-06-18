package srv

import (
	"fmt"
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"sync"
	"time"
)

var timerMap = make(map[int32]*timingwheel.Timer)

// RequestTracker
// @Description:
type RequestTracker struct {
	mu sync.Mutex

	pending map[int64]chan *exchange.Protocol
}

// Register
//
//	@Description: Register
//	@receiver rt
//	@param reqId
//	@return chan
func (rt *RequestTracker) Register(reqId int64) chan *exchange.Protocol {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	ch := make(chan *exchange.Protocol, 1)
	rt.pending[reqId] = ch
	return ch
}

// Complete 响应到达：投递数据并移除
func (rt *RequestTracker) Complete(reqId int64, resp *exchange.Protocol) {
	rt.mu.Lock()
	ch, ok := rt.pending[reqId]
	if ok {
		delete(rt.pending, reqId)
	}
	rt.mu.Unlock()

	if ok {
		ch <- resp
	}
}

// Remove 主动取消
func (rt *RequestTracker) Remove(reqId int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.pending, reqId)
}

// Transport
// @Description:
type Transport struct {

	//
	//  clients
	//  @Description: all client.
	//
	client *Client

	host string

	port int32

	config *configs.ClientConfig

	tracker *RequestTracker
}

// NewTransport
//
//	@Description: Init Transport.
//	@param ct
//	@return Transport
func NewTransport(config *configs.ClientConfig) *Transport {
	//start reconnection.
	return &Transport{
		host:   config.ServerHost,
		port:   config.ServerPort,
		config: config,
		tracker: &RequestTracker{
			pending: make(map[int64]chan *exchange.Protocol),
			//mu:      sync.Mutex{},
		},
	}
}

func (t *Transport) Connection(opts ...ClientOption) {
	t.client = NewClient(t.host, t.port)
	err := t.client.Connection("tcp", opts...)
	t.client.AddHandler(&CheckHandler{
		tracker: t.tracker,
	})
	//The error add to reconnection list.
	if err != nil {
		log.Warn("Connection to server error:%s", err)
		addChecking(t.client)
	} else {
		err := t.client.OpenTunnel("http")
		if err != nil {
			log.Warn("Connection to server error:%s", err)
		}
	}
}

func (t *Transport) WriteAsync(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	request, _ := exchange.NewRequest(message)
	ch := t.tracker.Register(request.ReqId)
	defer t.tracker.Remove(request.ReqId)
	t.client.cct.Write(request.Bytes())
	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

type ClientScheduler struct {
}

func (t *ClientScheduler) Next(t2 time.Time) time.Time {
	return t2.Add(3000 * time.Millisecond)
}

type CheckHandler struct {
	BaseClientHandler
	tracker *RequestTracker
}

func (b *CheckHandler) Close(cct *ClientControl) {
	addChecking(cct.cli)
}

func (b *CheckHandler) Read(r *exchange.Protocol, cct *ClientControl) error {
	//Heart info.
	if r.Cmd == exchange.Heart {
		log.Debug("Receiver PONG info: %S", cct.cli.getAddress())
		return nil
	} else {
		b.tracker.Complete(r.ReqId, r)
		return nil
	}
}

func (b *CheckHandler) Timeout(cct *ClientControl) {
	var h = exchange.Heartbeat{
		Value: "PING",
	}
	request, _ := exchange.NewRequest(h)
	cct.Write(request.Bytes())
}

func checking(cli *Client) {
	if !cli.IsConnection() {
		log.Warn("Connection %s Not Active, start reconnection.", cli.getAddress())
		err := cli.doConnection()
		if err != nil {
			log.Warn("Reconnection %s Fail, next time still running.", cli.getAddress())
		} else {
			log.Info("👍<--Reconnection %s success OK.✅-->", cli.getAddress())
		}
	}
	defer func() {
		if cli.IsConnection() {
			timer, ok := timerMap[cli.id]
			if ok {
				timer.Stop()
				delete(timerMap, cli.id)
			}
		}
	}()
}

func addChecking(cli *Client) {
	if _, ok := timerMap[cli.id]; ok {
		return
	}
	t := newWheel.ScheduleFunc(&ClientScheduler{}, func() {
		checking(cli)
	})
	timerMap[cli.id] = t
}
