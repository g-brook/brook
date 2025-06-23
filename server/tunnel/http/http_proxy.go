package http

import (
	"context"
	"errors"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	RouteInfoKey = "routeInfo"
	ProxyKey     = "proxy"
)

type RouteFunction func(request *http.Request) (net.Conn, error)
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
		BufferPool: utils.GetBuffPool32k(),
		Rewrite: func(request *httputil.ProxyRequest) {
			out := request.Out
			in := request.In
			out.Header["X-Forwarded-For"] = in.Header["X-Forwarded-For"]
			out.URL.Scheme = "http"
			out.URL.Host = out.Host
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				value := ctx.Value(RouteInfoKey)
				switch value.(type) {
				case error:
					return nil, value.(error)
				case net.Conn:
					return value.(net.Conn), nil
				}
				return nil, errors.New("not found path")
			},
			Proxy: func(req *http.Request) (*url.URL, error) {
				return req.URL, nil
			},
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			log.Warn("Not found path %v", err)
			writer.WriteHeader(http.StatusNotFound)
			_, _ = writer.Write(Get404Info())
		},
	}
	return &Proxy{
		proxy:    reverseProxy,
		routeFun: fun,
	}
}
