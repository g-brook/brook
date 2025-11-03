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

package http

import (
	"io"
	"sync"
	"time"

	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/ringbuffer"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
	"github.com/gobwas/ws"
)

type Future interface {
	Done(data []byte)
	ReqId() int64
	Close()
}

type WsFuture struct {
	buffer  *ringbuffer.RingBuffer
	reqId   int64
	tracker *HttpTracker
	isClose bool
}

func (f *WsFuture) Read(bytes []byte) (int, error) {
	if f.isClose {
		return 0, io.EOF
	}
	return f.buffer.Read(bytes)
}

func (f *WsFuture) ReqId() int64 {
	return f.reqId
}

func (f *WsFuture) Close() {
	f.isClose = true
}

func newWsFuture(tracker *HttpTracker, reqId int64) *WsFuture {
	buffer := ringbuffer.NewRingBuffer(1024)
	future := &WsFuture{
		buffer:  buffer,
		reqId:   reqId,
		tracker: tracker,
	}
	tracker.PutRequest(future)
	return future
}

func (f *WsFuture) Done(data []byte) {
	_, _ = f.buffer.Write(data)
}

type ResponseFuture struct {
	done    chan struct{}
	data    []byte
	err     error
	mu      sync.Mutex
	reqId   int64
	tracker *HttpTracker
}

func newResponseFuture(tracker *HttpTracker) *ResponseFuture {
	future := &ResponseFuture{
		done:    make(chan struct{}),
		reqId:   newReqId(),
		tracker: tracker,
	}
	tracker.PutRequest(future)
	return future
}

func (f *ResponseFuture) ReqId() int64 {
	return f.reqId
}

func (f *ResponseFuture) Done(data []byte) {
	f.mu.Lock()
	already := f.isDoneLocked()
	if !already {
		f.data = data
		f.Close()
	}
	f.mu.Unlock()
}

func (f *ResponseFuture) Error(err error) {
	f.mu.Lock()
	already := f.isDoneLocked()
	if !already {
		f.err = err
		f.Close()
	}
	f.mu.Unlock()
}

func (f *ResponseFuture) Close() {
	select {
	case <-f.done:
		return
	default:
		close(f.done)
	}
}

func (f *ResponseFuture) Wait() ([]byte, error) {
	<-f.done
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data, f.err
}

// WaitTimeout waits with timeout
func (f *ResponseFuture) WaitTimeout(d time.Duration) ([]byte, error) {
	select {
	case <-f.done:
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.data, f.err
	case <-time.After(d):
		return nil, timeoutErr
	}
}

func (f *ResponseFuture) isDoneLocked() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// HttpTracker httpx tracker
type HttpTracker struct {
	mu       sync.Mutex
	channel  transport.Channel
	trackers *hash.SyncMap[int64, Future]
}

func NewHttpTracker(channel transport.Channel) *HttpTracker {
	return &HttpTracker{
		trackers: hash.NewSyncMap[int64, Future](),
		channel:  channel,
	}
}

func (receiver *HttpTracker) GetFuture(reqId int64) (Future, bool) {
	return receiver.trackers.Load(reqId)
}

func (receiver *HttpTracker) Run() {
	threading.GoSafe(receiver.readRev)
}

func (receiver *HttpTracker) PutRequest(future Future) {
	receiver.trackers.Store(future.ReqId(), future)
}

func (receiver *HttpTracker) readRev() {
	readResponse := func() {
		pt := exchange.NewTunnelRead()
		err := pt.Read(receiver.channel)
		if err != nil {
			return
		}
		receiver.send(pt)
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

func (receiver *HttpTracker) send(pt *exchange.TunnelProtocol) {
	ch, ok := receiver.trackers.Load(pt.ReqId)
	if ok {
		if pt.Ver == exchange.WebsocketV2 {
			if pt.AttrLen > 0 {
				b := pt.Attr[0]
				if b == byte(ws.OpClose) {
					receiver.Close(pt.ReqId)
					return
				}
			}
		}
		ch.Done(pt.Data)
	}
	if pt.Ver == exchange.V1 || pt.Ver == exchange.WebsocketV1 {
		receiver.Close(pt.ReqId)
	}

}

func (receiver *HttpTracker) Close(reqId int64) {
	ch, ok := receiver.trackers.Load(reqId)
	if ok {
		ch.Close()
	}
	receiver.trackers.Delete(reqId)
}
