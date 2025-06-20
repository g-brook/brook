package clis

import (
	"fmt"
	"github.com/RussellLuo/timingwheel"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"sync"
	"time"
)

var timerMap = make(map[int32]*timingwheel.Timer)

// RequestTracker is will wait response for remote server and save request id of the map.
type RequestTracker struct {
	mu      sync.Mutex
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

// Complete delivers a response and removes the tracker entry.
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

// Remove   explicitly removes a request from tracker.
func (rt *RequestTracker) Remove(reqId int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.pending, reqId)
}

// Transport
// @Description:Transport manages client and request tracking.
type Transport struct {

	// client　is net connection.
	client *Client

	host string

	port int

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
		},
	}
}

func (t *Transport) Connection(opts ...ClientOption) {
	t.client = NewClient(t.host, t.port)
	err := t.client.Connection("tcp", opts...)
	t.client.AddHandler(&CheckHandler{
		tracker:   t.tracker,
		transport: t,
	})
	//The error add to reconnection list.
	if err != nil {
		log.Warn("Connection to server error:%s", err)
		addChecking(t)
	} else {
		t.openTunnel()
	}
}

func (t *Transport) openTunnel() {
	if t.client.isSmux() && t.config.Tunnels != nil {
		for _, cfg := range t.config.Tunnels {
			if err := t.client.OpenTunnel(cfg); err != nil {
				log.Warn("Connection to server error:%s %v", cfg.Type, err)
			}
		}
	}
}

func (t *Transport) WriteAsync(message exchange.InBound, timeout time.Duration) (*exchange.Protocol, error) {
	request, _ := exchange.NewRequest(message)
	ch := t.tracker.Register(request.ReqId)
	defer t.tracker.Remove(request.ReqId)
	err := t.client.cct.Write(request.Bytes())
	if err != nil {
		return nil, err
	}
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
	return t2.Add(3 * time.Second)
}

type CheckHandler struct {
	BaseClientHandler
	transport *Transport
	tracker   *RequestTracker
}

func (b *CheckHandler) Close(cct *ClientControl) {
	addChecking(b.transport)
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
	_ = cct.Write(request.Bytes())
}

func checking(tp *Transport) {
	cli := tp.client
	if !cli.IsConnection() {
		log.Warn("Connection %s Not Active, start reconnection.", cli.getAddress())
		err := cli.doConnection()
		if err != nil {
			log.Warn("Reconnection %s Fail, next time still running.", cli.getAddress())
		} else {
			log.Info("👍<--Reconnection %s success OK.✅-->", cli.getAddress())
			tp.openTunnel()
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

func addChecking(tp *Transport) {
	if _, ok := timerMap[tp.client.id]; ok {
		return
	}
	t := utils.NewWheel.ScheduleFunc(&ClientScheduler{}, func() {
		checking(tp)
	})
	timerMap[tp.client.id] = t
}
