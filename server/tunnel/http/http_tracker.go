package http

import (
	"bufio"
	"github.com/brook/common/log"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/brook/common/transport"
)

// HttpTracker http tracker
type HttpTracker struct {
	mu       sync.Mutex
	channel  transport.Channel
	trackers map[string]chan []byte
}

func NewHttpTracker(channel transport.Channel) *HttpTracker {
	return &HttpTracker{
		trackers: make(map[string]chan []byte),
		channel:  channel,
	}
}

func (receiver *HttpTracker) Run() {
	go receiver.readRev()
}

func (receiver *HttpTracker) AddRequest(reqId string) chan []byte {
	ch := make(chan []byte, 1)
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.trackers[reqId] = ch
	return ch
}

func (receiver *HttpTracker) readRev() {
	readResponse := func() {
		response, err := http.ReadResponse(bufio.NewReader(receiver.channel.GetReader()), nil)
		if err != nil {
			log.Error("read response error", err)
			return
		}
		header, err := httputil.DumpResponse(response, true)
		if err != nil {
			log.Error("read response error", err)
			return
		}
		reqId := response.Header.Get(RequestInfoKey)
		receiver.send(reqId, header)
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

func (receiver *HttpTracker) send(reqId string, buffer []byte) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	ch := receiver.trackers[reqId]
	if ch != nil {
		ch <- buffer
	}
	delete(receiver.trackers, reqId)

}

func (receiver *HttpTracker) Close(reqId string) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	c := receiver.trackers[reqId]
	if c != nil {
		close(c)
	}
	delete(receiver.trackers, reqId)
}
