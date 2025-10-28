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
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brook/common/exchange"
	"github.com/brook/common/httpx"
	"github.com/brook/common/iox"
	"github.com/brook/common/log"
	"golang.org/x/net/websocket"
)

var (
	RouteInfoKey   = "routeInfo"
	RequestInfoKey = "httpRequestId"
	ProxyKey       = "proxy"
	ForwardedKey   = "X-Forwarded-For"
	index          atomic.Int64
)

type timeoutError struct {
}

func (timeoutError) Error() string   { return "read timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true } // optional

type ProxyConnection struct {
	net.Conn
	tracker *HttpTracker
	reqId   int64
	readBuf []byte
	readCh  chan []byte
	mu      sync.Mutex
}

func NewProxyConnection(conn net.Conn, reqId int64, tracker *HttpTracker) *ProxyConnection {
	return &ProxyConnection{
		Conn:    conn,
		reqId:   reqId,
		tracker: tracker,
		readCh:  tracker.trackers[reqId],
	}
}

func (proxy *ProxyConnection) Write(b []byte) (n int, err error) {
	//encode to tunnel.
	request := exchange.NewTunnelWriter(b, proxy.reqId)
	err = request.Writer(proxy.Conn)
	return len(b), err
}

func (proxy *ProxyConnection) Read(p []byte) (n int, err error) {
	if len(proxy.readBuf) > 0 {
		n := copy(p, proxy.readBuf)
		proxy.readBuf = proxy.readBuf[n:]
		return n, nil
	}
	select {
	case data, ok := <-proxy.readCh:
		if !ok {
			return 0, io.EOF
		}
		n = copy(p, data)
		if n < len(data) {
			proxy.readBuf = append(proxy.readBuf, data[n:]...)
		}
		return n, nil
	case <-time.After(time.Second * 5):
		return 0, timeoutError{}
	}
}

func (proxy *ProxyConnection) Close() error {
	proxy.tracker.Close(proxy.reqId)
	return nil
}

type Proxy struct {
	httpProxy http.Handler
	websocket func(ctx context.Context) websocket.Handler
	routeFun  RouteFunction
}

func (h *Proxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newCtx := request.Context()
	newCtx = context.WithValue(newCtx, ProxyKey, true)
	if info, err := h.routeFun(request); err != nil {
		newCtx = context.WithValue(newCtx, RouteInfoKey, err)
	} else {
		reqId := index.Add(1)
		newCtx = context.WithValue(newCtx, RequestInfoKey, reqId)
		newCtx = context.WithValue(newCtx, RouteInfoKey, info)
	}
	newReq := request.Clone(newCtx)
	if isWebSocket(request) {
		h.websocket(newCtx).ServeHTTP(writer, request)
	} else {
		h.httpProxy.ServeHTTP(writer, newReq)
	}
}

func isWebSocket(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Connection")) == "upgrade" &&
		strings.ToLower(r.Header.Get("Upgrade")) == "websocket"
}

func NewHttpProxy(fun RouteFunction) *Proxy {
	return &Proxy{
		httpProxy: httpProxy(),
		websocket: websocketProxy,
		routeFun:  fun,
	}
}

func websocketProxy(ctx context.Context) websocket.Handler {
	workFunction := func(conn, workConn net.Conn) {
		errors := iox.Pipe(conn, workConn)
		if len(errors) > 0 {
			log.Warn("copy error %v", errors)
		}
	}
	return func(conn *websocket.Conn) {
		value := ctx.Value(RouteInfoKey)
		id := ctx.Value(RequestInfoKey)
		switch v := value.(type) {
		case error:
			return
		case *RouteInfo:
			targert, err := v.getProxyConnection(v.httpId, id.(int64))
			if err != nil {
				return
			}
			workFunction(conn, targert)
		}
	}
}

func httpProxy() *httputil.ReverseProxy {
	reverseProxy := &httputil.ReverseProxy{
		BufferPool: iox.GetBytePool32k(),
		Rewrite: func(request *httputil.ProxyRequest) {
			out := request.Out
			in := request.In
			out.Header[ForwardedKey] = in.Header[ForwardedKey]
			out.Header[RequestInfoKey] = in.Header[RequestInfoKey]
			out.URL.Scheme = "http"
			out.URL.Host = out.Host
		},
		ModifyResponse: func(response *http.Response) error {
			response.Header.Del(RequestInfoKey)
			return nil
		},
		Transport: &http.Transport{
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
			MaxIdleConnsPerHost:   0,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				value := ctx.Value(RouteInfoKey)
				id := ctx.Value(RequestInfoKey)
				switch v := value.(type) {
				case error:
					return nil, v
				case *RouteInfo:
					return v.getProxyConnection(v.httpId, id.(int64))
				}
				return nil, nil
			},
			Proxy: func(req *http.Request) (*url.URL, error) {
				return req.URL, nil
			},
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			state := http.StatusOK
			if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
				state = http.StatusGatewayTimeout
			} else {
				state = http.StatusNotFound
			}
			log.Error("Not found path %v", err)
			writer.WriteHeader(state)
			_, _ = writer.Write(httpx.GetPageNotFound(state))
		},
	}
	return reverseProxy
}
