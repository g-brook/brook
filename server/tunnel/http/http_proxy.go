package http

import (
	"context"
	io2 "github.com/brook/common/aio"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	RouteInfoKey = "routeInfo"
	ProxyKey     = "proxy"
)

type ProxyConnection struct {
	transport.Channel
}

func NewProxyConnection(conn transport.Channel) *ProxyConnection {
	return &ProxyConnection{
		Channel: conn,
	}
}

func (proxy *ProxyConnection) Close() error {
	//  close
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
			out.Header["X-Forwarded-For"] = in.Header["X-Forwarded-For"]
			out.URL.Scheme = "http"
			out.URL.Host = out.Host
		},
		ModifyResponse: func(response *http.Response) error {
			return nil
		},
		Transport: &http.Transport{
			ResponseHeaderTimeout: 20 * time.Second,
			IdleConnTimeout:       60 * time.Second,
			MaxIdleConnsPerHost:   5,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				value := ctx.Value(RouteInfoKey)
				switch v := value.(type) {
				case error:
					return nil, v
				case *RouteInfo:
					return v.getProxyConnection(v.proxyId)
				}
				return nil, nil
			},
			Proxy: func(req *http.Request) (*url.URL, error) {
				return req.URL, nil
			},
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			//if errors.Is(err, io.EOF) {
			//	fmt.Println("发送了错误...")
			//}
			log.Error("Not found path %v", err)
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write(utils.GetPageNotFound())
		},
	}
	return &Proxy{
		proxy:    reverseProxy,
		routeFun: fun,
	}
}
