package http

import (
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/transport"
)

// HttpTracker http tracker
type HttpTracker struct {
	mu       sync.Mutex
	channel  transport.Channel
	trackers map[int64]chan []byte
}

func NewHttpTracker(channel transport.Channel) *HttpTracker {
	return &HttpTracker{
		trackers: make(map[int64]chan []byte),
		channel:  channel,
	}
}

func (receiver *HttpTracker) Run() {
	go receiver.readRev()
}

func (receiver *HttpTracker) AddRequest(reqId int64) chan []byte {
	ch := make(chan []byte, 1)
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.trackers[reqId] = ch
	return ch
}

func (receiver *HttpTracker) readRev() {
	readResponse := func() {
		pt := exchange.NewTunnelRead()
		err := pt.Read(receiver.channel)
		if err != nil {
			return
		}
		reqId := pt.ReqId
		receiver.send(reqId, pt.Data)
	}
	for {
		select {
		case <-receiver.channel.Done():
			return
		default:
		}
		readResponse()
	}
}

func (receiver *HttpTracker) send(reqId int64, buffer []byte) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	ch := receiver.trackers[reqId]
	if ch != nil {
		ch <- buffer
	}
	delete(receiver.trackers, reqId)

}

func (receiver *HttpTracker) Close(reqId int64) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	c := receiver.trackers[reqId]
	if c != nil {
		close(c)
	}
	delete(receiver.trackers, reqId)
}
