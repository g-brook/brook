package exchange

import (
	"fmt"
	"sync"
	"time"
)

// RequestTracker is will wait response for remote server and save request id of the map.

var Tracker *RequestTracker

func init() {
	Tracker = &RequestTracker{
		pending: make(map[int64]chan *Protocol),
	}
}

type RequestTracker struct {
	mu      sync.Mutex
	pending map[int64]chan *Protocol
}

// Register
//
//	@Description: Register
//	@receiver rt
//	@param reqId
//	@return chan
func (rt *RequestTracker) Register(reqId int64) chan *Protocol {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	ch := make(chan *Protocol, 1)
	rt.pending[reqId] = ch
	return ch
}

// Complete delivers a response and removes the tracker entry.
func (rt *RequestTracker) Complete(resp *Protocol) bool {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	ch, ok := rt.pending[resp.ReqId]
	if ok {
		delete(rt.pending, resp.ReqId)
		ch <- resp
	}
	return ok
}

func SyncWriteInBound(message InBound, timeout time.Duration, writer func(protocol *Protocol) error) (*Protocol, error) {
	request, _ := NewRequest(message)
	return SyncWriteByProtocol(request, timeout, writer)
}

func SyncWriteByProtocol(message *Protocol, timeout time.Duration, writer func(protocol *Protocol) error) (*Protocol, error) {
	ch := Tracker.Register(message.ReqId)
	defer Tracker.Remove(message.ReqId)
	err := writer(message)
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
