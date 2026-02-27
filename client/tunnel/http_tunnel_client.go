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

package tunnel

import (
	"context"
	"errors"
	"io"

	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/common/transport"
	"github.com/gobwas/ws"
)

func NewHttpTunnelClient(config *configs.ClientTunnelConfig) (*HttpTunnelClient, error) {
	if config.HttpId == "" {
		log.Warn("HttpId is empty,http tunnel client will not connect")
		return nil, errors.New("HttpId is empty,http tunnel client will not connect")
	}
	tunnelClient := clis.NewBaseTunnelClient(config, true)
	client := HttpTunnelClient{
		BaseTunnelClient: tunnelClient,
		websocket:        newWebsocketClientManager(),
		http:             NewHttpClientManager(),
	}
	tunnelClient.DoOpen = client.initOpen
	return &client, nil
}

// HttpTunnelClient is a tunnel client that handles HTTP connections.
type HttpTunnelClient struct {
	*clis.BaseTunnelClient
	websocket *WebsocketClientManager
	http      *HttpClientManager
}

// GetName returns the name of the tunnel client.
func (h *HttpTunnelClient) GetName() string {
	return "HttpTunnelClient"
}

// initOpen initializes the HTTP tunnel client by registering it and logging the result.
// Parameters:
//   - stream: The smux stream to use.
//
// Returns:
//   - error: An error if the registration fails.
func (h *HttpTunnelClient) initOpen(*transport.SChannel) error {
	err := h.AsyncRegister(h.GetRegisterReq(), func(p *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
		if p.IsSuccess() {
			var finnish = make(chan int)
			threading.GoSafe(func() {
				err := h.bindHandler(p, rw, ctx)
				if err != nil {
					finnish <- 1
				}
			})
			rsp, _ := exchange.Parse[exchange.RegisterReqAndRsp](p.Data)
			err := h.OpenWorkerToManager(rsp)
			if err != nil {
				return exchange.CloseError
			}
			<-finnish
			log.Debug("Exit handler......%s:%s", rsp.ProxyId, rsp.HttpId)
			return nil
		}
		log.Error("Connection local address success then Client to server register fail:%v", p.RspMsg)
		return exchange.CloseError
	})
	return err
}

// bindHandler handles the binding of HTTP tunnel client requests
func (h *HttpTunnelClient) bindHandler(p *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
	if !p.IsSuccess() {
		return exchange.CloseError
	}
	cwc, _ := exchange.Parse[exchange.ClientWorkConnReq](p.Data)
	log.Info("Open worker success. %s:%s:%v", cwc.ProxyId, cwc.HttpId, cwc.TunnelPort)
	loopRead := func() error {
		pt := exchange.NewTunnelRead()
		err := pt.Read(rw)
		if err != nil {
			return err
		}
		// If the pt.Ver is v1, that it is a http request
		if pt.Ver == exchange.V1 || pt.Ver == exchange.WebsocketV1 {
			isWs := pt.Ver == exchange.WebsocketV1
			httpBridge, err := h.http.GetHttpBridge(ctx, rw, h.GetCfg().Destination, pt.ReqId, isWs)
			if err != nil {
				log.Warn("GetHttpBridge fail %v", err)
				response := getErrorResponse(httpError)
				return exchange.NewTunnelWriter(response, pt.ReqId).Writer(rw)
			}
			//to websocket ss.
			if isWs {
				path := string(pt.Attr)
				httpBridge.upgrader(func() {
					wsBridge, err := httpBridge.websocket(ctx, path, h.websocket)
					if err != nil {
						h.websocket.closeWebsocketLeft(rw, pt.ReqId)
						return
					}
					wsBridge.toRunning()
					log.Debug("Connect to websocket success.")
					return
				})
			}
			_, _ = httpBridge.WriterToRight(pt.Data)
		} else if pt.Ver == exchange.WebsocketV2 {
			websocketCall(pt, rw, h.websocket)
		}
		return nil

	}
	// Main loop to handle incoming requests
	for {
		select {
		// Check for context cancellation
		case <-ctx.Done():
			return exchange.CloseError
		default:
		}
		// Process next request
		err := loopRead()
		if err == io.EOF {
			log.Debug("http stream close.")
			_ = rw.Close()
			return exchange.CloseError
		}
	}

}

func websocketCall(protocol *exchange.TunnelProtocol, left io.ReadWriteCloser, manager *WebsocketClientManager) {
	if len(protocol.Attr) <= 1 {
		return
	}
	opCode := ws.OpCode(protocol.Attr[0])
	path := string(protocol.Attr[1:])
	conn, b := manager.getWebsocketBridge(path)
	if !b {
		manager.closeWebsocketLeft(left, protocol.ReqId)
		return
	}
	conn.WriteToRight(opCode, protocol.Data)
}
