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
	"net/http"

	"github.com/brook/common/httpx"
	"github.com/brook/common/iox"
	"github.com/brook/common/log"
	"golang.org/x/net/websocket"
)

type WebsocketProxy struct {
	routeFun  RouteFunction
	websocket func(info *RouteInfo, writer http.ResponseWriter, request *http.Request) websocket.Handler
}

func NewWebsocketProxy(route func(req *http.Request) (*RouteInfo, error)) *WebsocketProxy {
	return &WebsocketProxy{
		websocket: websocketProxy,
		routeFun:  route,
	}
}

func websocketProxy(info *RouteInfo, writer http.ResponseWriter, request *http.Request) websocket.Handler {
	workFunction := func(websocketConnection *websocket.Conn, targetConn *ProxyConnection, reqId int64) {
		targetConn.isWebsocket = true
		targetConn.path = request.URL.Path
		request.URL.Scheme = "http"

		err := request.Write(targetConn)
		if err != nil {
			writer.WriteHeader(http.StatusBadGateway)
			_, _ = writer.Write([]byte("error"))
			_ = websocketConnection.Close()
			return
		}
		errors := iox.Pipe(websocketConnection, targetConn.websocket(websocketConnection.PayloadType))
		if len(errors) > 0 {
			log.Warn("copy error %v", errors)
		}
	}
	return func(conn *websocket.Conn) {
		targetConn, err := info.getProxyConnection(info.httpId)
		if err != nil {
			log.Error("get proxy connection error %v", err)
			return
		}
		id := newReqId()
		switch v := targetConn.(type) {
		case *ProxyConnection:
			workFunction(conn, v, id)
		default:
			_ = targetConn.Close()
		}
	}
}

func (h *WebsocketProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	info, err := h.routeFun(request)
	if err != nil {
		log.Error("route error %v", err)
		http.NotFound(writer, request)
		return
	}
	if isWebSocket(request) {
		h.websocket(info, writer, request).ServeHTTP(writer, request)
	} else {
		http.NotFound(writer, request)
	}
}

func isWebSocket(r *http.Request) bool {
	return httpx.IsWebSocketRequest(r)
}
