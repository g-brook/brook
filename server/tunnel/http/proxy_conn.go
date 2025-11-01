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
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brook/common/exchange"
	"github.com/gobwas/ws"
)

var (
	RouteInfoKey   = "routeInfo"
	RequestInfoKey = "httpRequestId"
	ProxyKey       = "httpProxy"
	ForwardedKey   = "X-Forwarded-For"
	index          atomic.Int64
	timeoutErr     = &timeoutError{}
)

func newReqId() int64 {
	return index.Add(1)
}

type timeoutError struct {
}

func (timeoutError) Error() string   { return "read timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true } // optional
type WebsocketConnection struct {
	net.Conn
	tracker     *HttpTracker
	future      *WsFuture
	path        string
	payloadType byte
}

func (wc *WebsocketConnection) Close() error {
	bytes := []byte(wc.path)
	attr := make([]byte, len(bytes)+1)
	attr[0] = byte(ws.OpClose)
	copy(attr[1:], bytes)
	writer := exchange.NewTunnelWebsocketWriterV2([]byte{}, attr, wc.future.ReqId())
	return writer.Writer(wc.Conn)
}

func (wc *WebsocketConnection) Write(b []byte) (n int, err error) {
	bytes := []byte(wc.path)
	attr := make([]byte, len(bytes)+1)
	attr[0] = wc.payloadType
	copy(attr[1:], bytes)
	writer := exchange.NewTunnelWebsocketWriterV2(b, attr, wc.future.ReqId())
	err = writer.Writer(wc.Conn)
	if err == io.EOF {
		wc.tracker.Close(wc.future.ReqId())
	}
	return len(b), err
}

func (wc *WebsocketConnection) Read(p []byte) (n int, err error) {
	for {
		read, err := wc.future.Read(p)
		if err == io.EOF {
			return 0, err
		}
		if read <= 0 {
			runtime.Gosched()
			continue
		}
		return read, err
	}
}

type ProxyConnection struct {
	net.Conn
	tracker     *HttpTracker
	readBuf     []byte
	future      *ResponseFuture
	mu          sync.Mutex
	isWebsocket bool
	path        string
}

func NewProxyConnection(conn net.Conn,
	tracker *HttpTracker) *ProxyConnection {
	return &ProxyConnection{
		Conn:    conn,
		tracker: tracker,
		future:  newResponseFuture(tracker),
	}
}

func (proxy *ProxyConnection) Write(b []byte) (n int, err error) {
	//encode to tunnel.
	var writer *exchange.TunnelProtocol
	id := proxy.future.ReqId()
	if proxy.isWebsocket {
		writer = exchange.NewTunnelWebsocketWriterV1(b, []byte(proxy.path), id)
	} else {
		writer = exchange.NewTunnelWriter(b, id)
	}
	err = writer.Writer(proxy.Conn)
	if err != nil {
		proxy.tracker.Close(id)
	}
	return len(b), err
}

func (proxy *ProxyConnection) Read(p []byte) (n int, err error) {
	if len(proxy.readBuf) > 0 {
		n := copy(p, proxy.readBuf)
		proxy.readBuf = proxy.readBuf[n:]
		return n, nil
	}
	bytes, err := proxy.future.WaitTimeout(5 * time.Second)
	if err != nil {
		return 0, err
	}
	n = copy(p, bytes)
	return n, nil
	//select {
	//case data, ok := <-proxy.future.Wait():
	//	if !ok {
	//		return 0, io.EOF
	//	}
	//	n = copy(p, data)
	//	if n < len(data) {
	//		proxy.readBuf = append(proxy.readBuf, data[n:]...)
	//	}
	//	return n, nil
	//case <-time.After(time.Second * 5):
	//	return 0, timeoutError{}
	//}
}

func (proxy *ProxyConnection) websocket(payloadType byte) net.Conn {
	return &WebsocketConnection{
		payloadType: payloadType,
		Conn:        proxy.Conn,
		tracker:     proxy.tracker,
		future:      newWsFuture(proxy.tracker, proxy.future.ReqId()),
		path:        proxy.path}
}

func (proxy *ProxyConnection) Close() error {
	if proxy.future != nil {
		proxy.tracker.Close(proxy.future.reqId)
	}
	return nil
}
