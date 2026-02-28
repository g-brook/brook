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
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/g-brook/brook/common/httpx"
	"github.com/g-brook/brook/common/iox"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/scmd/web/logger"
)

type Proxy struct {
	http     http.Handler
	routeFun RouteFunction
}

func (h *Proxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newCtx := request.Context()
	newCtx = context.WithValue(newCtx, ProxyKey, true)
	info, err := h.routeFun(request)
	if err != nil {
		newCtx = context.WithValue(newCtx, RouteInfoKey, err)
	} else {
		newCtx = context.WithValue(newCtx, RequestInfoKey, newReqId())
		newCtx = context.WithValue(newCtx, RouteInfoKey, info)
	}
	h.initHeader(writer, request, info)
	newReq := request.Clone(newCtx)
	h.http.ServeHTTP(writer, newReq)
}

func (h *Proxy) initHeader(writer http.ResponseWriter, request *http.Request, r *RouteInfo) {
	if writer.Header() != nil {
		header := writer.Header()
		header.Set("Connection", request.Header.Get("Connection"))
	}
	if r != nil {
		request.Header.Set(RequestHttpIdKey, r.httpId)
		request.Header.Set(RequestDomainKey, r.domain)
	}
}

func NewHttpProxy(fun RouteFunction, proxyId string) *Proxy {
	return &Proxy{
		http:     httpProxy(proxyId),
		routeFun: fun,
	}
}

func httpProxy(proxyId string) *httputil.ReverseProxy {
	reverseProxy := &httputil.ReverseProxy{
		BufferPool: iox.GetBytePool32k(),
		Rewrite: func(request *httputil.ProxyRequest) {
			out := request.Out
			in := request.In
			out.Header[ForwardedKey] = in.Header[ForwardedKey]
			out.Header[RequestInfoKey] = in.Header[RequestInfoKey]
			out.Header[RequestHttpIdKey] = in.Header[RequestHttpIdKey]
			out.URL.Scheme = "http"
			out.URL.Host = out.Host
		},
		ModifyResponse: func(response *http.Response) error {
			req := response.Request
			logger.WithWebLog(&logger.WebLogger{
				Protocol: req.Proto,
				Path:     req.URL.Path,
				Host:     req.Host,
				Method:   req.Method,
				ProxyId:  proxyId,
				Status:   response.StatusCode,
				HttpId:   req.Header.Get(RequestHttpIdKey),
				Time:     time.Now(),
			})
			response.Header.Del(RequestInfoKey)
			return nil
		},

		Transport: &http.Transport{
			ResponseHeaderTimeout: 5 * time.Second,
			DisableKeepAlives:     true,
			MaxIdleConnsPerHost:   0,
			IdleConnTimeout:       5 * time.Second,
			MaxIdleConns:          100,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				value := ctx.Value(RouteInfoKey)
				switch v := value.(type) {
				case error:
					return nil, v
				case *RouteInfo:
					connection, err := v.getProxyConnection(v.httpId)
					if err != nil {
						log.Error("get proxy connection error %v", err)
						return nil, err
					}
					return connection, err
				}
				return nil, nil
			},
			Proxy: func(req *http.Request) (*url.URL, error) {
				return req.URL, nil
			},
		},
		ErrorHandler: func(writer http.ResponseWriter, req *http.Request, err error) {
			if errors.Is(err, readDone) {
				return
			}
			state := http.StatusOK
			if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
				state = http.StatusGatewayTimeout
			} else {
				state = http.StatusNotFound
			}
			log.Error("Not found path %v", err)
			writer.WriteHeader(state)
			_, _ = writer.Write(httpx.GetPageNotFound(state))
			logger.WithWebLog(&logger.WebLogger{
				Protocol: req.Proto,
				Path:     req.URL.Path,
				Host:     req.Host,
				Method:   req.Method,
				ProxyId:  proxyId,
				Status:   state,
				HttpId:   req.Header.Get(RequestHttpIdKey),
				Time:     time.Now(),
			})
		},
	}
	return reverseProxy
}
