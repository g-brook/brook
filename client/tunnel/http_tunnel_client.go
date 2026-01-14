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

	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	"github.com/brook/common/transport"
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
	h.BaseTunnelClient.AddReadHandler(exchange.WorkerConnReq, h.bindHandler)
	rsp, err := h.Register(h.GetRegisterReq())
	if err != nil {
		log.Error("Register fail %v", err)
		return err
	}
	log.Info("Register success:PORT-%v", rsp.TunnelPort)
	return nil
}

// bindHandler handles the binding of HTTP tunnel client requests
func (h *HttpTunnelClient) bindHandler(_ *exchange.Protocol, rw io.ReadWriteCloser, ctx context.Context) error {
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
			return nil
		default:
		}
		// Process next request
		err := loopRead()
		if err == io.EOF {
			log.Debug("http stream close.")
			_ = rw.Close()
			return nil
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
