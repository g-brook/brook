package http

import (
	"bufio"
	"bytes"
	"github.com/brook/common/transport"
	"net/http"
	"sync"
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
	receiver.trackers[reqId] = ch
	return ch
}

func (receiver *HttpTracker) readRev() {
	readResponse := func() {
		response, err := http.ReadResponse(bufio.NewReader(receiver.channel.GetReader()), nil)
		if err != nil {
			return
		}
		buffer := bytes.NewBuffer(nil)
		err = response.Write(buffer)
		if err != nil {
			return
		}
		reqId := response.Header.Get(RequestInfoKey)
		ch := receiver.trackers[reqId]
		if ch != nil {
			ch <- buffer.Bytes()
		}
		delete(receiver.trackers, reqId)
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

func (receiver *HttpTracker) Close(reqId string) {
	delete(receiver.trackers, reqId)
}
