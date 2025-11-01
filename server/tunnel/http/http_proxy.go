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
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/brook/common/httpx"
	"github.com/brook/common/iox"
	"github.com/brook/common/log"
)

type HttpProxy struct {
	http     http.Handler
	routeFun RouteFunction
}

func (h *HttpProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	newCtx := request.Context()
	newCtx = context.WithValue(newCtx, ProxyKey, true)
	if info, err := h.routeFun(request); err != nil {
		newCtx = context.WithValue(newCtx, RouteInfoKey, err)
	} else {
		newCtx = context.WithValue(newCtx, RequestInfoKey, newReqId())
		newCtx = context.WithValue(newCtx, RouteInfoKey, info)
	}
	newReq := request.Clone(newCtx)
	h.http.ServeHTTP(writer, newReq)
}

func NewHttpProxy(fun RouteFunction) *HttpProxy {
	return &HttpProxy{
		http:     httpProxy(),
		routeFun: fun,
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
					connection, err := v.getProxyConnection(v.httpId, id.(int64))
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
