package http

import (
	"context"
	io2 "github.com/brook/common/aio"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	RouteInfoKey = "routeInfo"

	RequestInfoKey = "httpRequestId"

	ProxyKey = "proxy"

	ForwardedKey = "X-Forwarded-For"
)

type timeoutError struct{}

func (timeoutError) Error() string   { return "read timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true } // optional

type ProxyConnection struct {
	net.Conn
	tracker *HttpTracker
	reqId   string
	readBuf []byte
	readCh  chan []byte
}

func NewProxyConnection(conn net.Conn, reqId string, tracker *HttpTracker) *ProxyConnection {
	return &ProxyConnection{
		Conn:    conn,
		reqId:   reqId,
		tracker: tracker,
		readCh:  tracker.trackers[reqId],
	}
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
	proxy    http.Handler
	routeFun RouteFunction
}

func (h *Proxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newCtx := request.Context()
	newCtx = context.WithValue(newCtx, ProxyKey, true)
	if info, err := h.routeFun(request); err != nil {
		newCtx = context.WithValue(newCtx, RouteInfoKey, err)
	} else {
		reqId := uuid.NewString()
		request.Header.Set(RequestInfoKey, reqId)
		newCtx = context.WithValue(newCtx, RequestInfoKey, reqId)
		newCtx = context.WithValue(newCtx, RouteInfoKey, info)
	}
	newReq := request.Clone(newCtx)
	h.proxy.ServeHTTP(writer, newReq)
}

func NewHttpProxy(fun RouteFunction) *Proxy {
	reverseProxy := &httputil.ReverseProxy{
		BufferPool: io2.GetBuffPool32k(),
		Rewrite: func(request *httputil.ProxyRequest) {
			out := request.Out
			in := request.In
			out.Header[ForwardedKey] = in.Header[ForwardedKey]
			out.Header[RequestInfoKey] = in.Header[RequestInfoKey]
			out.URL.Scheme = "http"
			out.URL.Host = out.Host
		},
		ModifyResponse: func(response *http.Response) error {
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
					return v.getProxyConnection(v.proxyId, id.(string))
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
			_, _ = writer.Write(utils.GetPageNotFound(state))
		},
	}
	return &Proxy{
		proxy:    reverseProxy,
		routeFun: fun,
	}
}
