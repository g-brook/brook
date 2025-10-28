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
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/threading"
	"github.com/brook/common/transport"
)

// HttpTracker httpx tracker
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
	threading.GoSafe(receiver.readRev)
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
