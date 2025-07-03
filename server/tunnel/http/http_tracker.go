package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
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

//// 分块读取HTTP响应
//func (receiver *HttpTracker) readRev() {
//	reader := bufio.NewReader(receiver.channel.GetReader())
//	for {
//		select {
//		case <-receiver.channel.Done():
//			return
//		default:
//		}
//		// 读取响应头
//		response, err := http.ReadResponse(reader, nil)
//		if err != nil {
//			fmt.Println("读取响应头错误:", err)
//			return
//		}
//
//		reqId := response.Header.Get(RequestInfoKey)
//		buffer := bytes.NewBuffer(nil)
//
//		// 写入响应头
//		err = response.Header.Write(buffer)
//		if err != nil {
//			fmt.Println("写入响应头错误:", err)
//			return
//		}
//		buffer.WriteString("\r\n")
//
//		// 分块读取响应体
//		if response.Body != nil {
//			buf := make([]byte, 4096)
//			for {
//				n, err := response.Body.Read(buf)
//				if n > 0 {
//					buffer.Write(buf[:n])
//				}
//				if err != nil {
//					break
//				}
//			}
//			response.Body.Close()
//		}
//
//		// 发送完整响应
//		receiver.send(reqId, buffer)
//	}
//}

func (receiver *HttpTracker) readRev() {
	readResponse := func() {
		response, err := http.ReadResponse(bufio.NewReader(receiver.channel.GetReader()), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		buffer := bytes.NewBuffer(nil)
		err = response.Write(buffer)
		if err != nil {
			return
		}
		reqId := response.Header.Get(RequestInfoKey)
		receiver.send(reqId, buffer)
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

func (receiver *HttpTracker) send(reqId string, buffer *bytes.Buffer) {
	receiver.mu.Lock()
	ch := receiver.trackers[reqId]
	ibytes := buffer.Bytes()
	fmt.Println(string(ibytes))
	if ch != nil {
		ch <- ibytes
	}
	delete(receiver.trackers, reqId)
	receiver.mu.Unlock()
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
