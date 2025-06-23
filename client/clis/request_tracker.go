package clis

import (
	"fmt"
	"github.com/brook/common/exchange"
	"sync"
	"time"
)

// RequestTracker is will wait response for remote server and save request id of the map.

var Tracker *RequestTracker

func init() {
	Tracker = &RequestTracker{
		pending: make(map[int64]chan *exchange.Protocol),
	}
}

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
	defer rt.mu.Unlock()
	ch, ok := rt.pending[reqId]
	if ok {
		delete(rt.pending, reqId)
		ch <- resp
	}
}

func SyncWrite(message exchange.InBound, timeout time.Duration, writer func([]byte) error) (*exchange.Protocol, error) {
	request, _ := exchange.NewRequest(message)
	ch := Tracker.Register(request.ReqId)
	defer Tracker.Remove(request.ReqId)
	err := writer(request.Bytes())
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

// Remove   explicitly removes a request from tracker.
func (rt *RequestTracker) Remove(reqId int64) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	delete(rt.pending, reqId)
}
