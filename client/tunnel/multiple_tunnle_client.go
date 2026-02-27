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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/g-brook/brook/client/cli"
	"github.com/g-brook/brook/client/clis"
	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/xtaci/smux"
)

var (
	globalMultipleClient *MultipleTunnelClient
	initOnce             sync.Once
)

const (
	// HTTP隧道预创建的工作连接数
	httpWorkerConnCount = 2
)

// This function is used to register a new tunnel client for the TCP protocol
func init() {
	ft := func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		initOnce.Do(func() {
			globalMultipleClient = &MultipleTunnelClient{
				sessions: hash.NewSyncMap[string, *smux.Session](),
			}
			globalMultipleClient.messageListener()
		})
		return &tunnelClientWrapper{
			multiClient: globalMultipleClient,
			config:      config,
		}
	}
	clis.RegisterTunnelClient(lang.Tcp, ft)
	clis.RegisterTunnelClient(lang.Udp, ft)
	clis.RegisterTunnelClient(lang.Http, ft)
	clis.RegisterTunnelClient(lang.Https, ft)
}

type MultipleTunnelClient struct {
	sessions  *hash.SyncMap[string, *smux.Session]
	closeOnce sync.Once
}

type tunnelClientWrapper struct {
	multiClient *MultipleTunnelClient
	config      *configs.ClientTunnelConfig
}

func (w *tunnelClientWrapper) Done() <-chan struct{} {
	return nil
}

func (w *tunnelClientWrapper) GetName() string {
	return "tunnel-client-wrapper"
}

func (m *MultipleTunnelClient) messageListener() {
	clis.ManagerTransport.AddMessageNotify(exchange.WorkerConnReq, func(r *exchange.Protocol) error {
		threading.GoSafe(func() {
			reqWorker, err := exchange.Parse[exchange.WorkConnReq](r.Data)
			if err != nil {
				log.Error("Parse WorkConnReq error: %v", err)
				return
			}
			config := clis.ManagerTransport.GetConfig(reqWorker.ProxyId)
			if config == nil {
				log.Warn("configs is nil %v", reqWorker.ProxyId)
				return
			}
			id := config.ProxyId
			session, b := m.sessions.Load(id)
			if !b {
				log.Warn("not found session %v", reqWorker.ProxyId)
				return
			}
			client, err := newTunnelClient(config, m)
			if err != nil {
				log.Error("newTunnelClient error: %v", err)
				return
			}
			_ = client.Open(session)
		})
		return nil
	})
}

func newTunnelClient(config *configs.ClientTunnelConfig, m *MultipleTunnelClient) (clis.TunnelClient, error) {
	switch config.TunnelType {
	case lang.Tcp:
		return NewTcpTunnelClient(config, m)
	case lang.Udp:
		return NewUdpTunnelClient(config, m)
	case lang.Http, lang.Https:
		return NewHttpTunnelClient(config)
	}
	return nil, errors.New("unknown tunnel type")
}

// Open This function opens a TCP tunnel server for a given session.
func (w *tunnelClientWrapper) Open(session *smux.Session) error {
	// Create a new OpenTunnelReq struct with the proxy ID, tunnel type, and tunnel port.
	req := &exchange.OpenTunnelReq{
		ProxyId: w.config.ProxyId,
		UnId:    clis.ManagerTransport.UnId,
	}
	rsp, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		log.Error("Open tunnel server error: %v", err)
		return err
	}
	if !rsp.IsSuccess() {
		log.Error("Open %v tunnel server error %v", w.config.ProxyId, rsp.RspMsg)
		return fmt.Errorf("open tunnel failed: %s", rsp.RspMsg)
	}
	rspObj, err := exchange.Parse[exchange.OpenTunnelResp](rsp.Data)
	if err != nil {
		log.Error("Parse OpenTunnelResp error: %v", err)
		return fmt.Errorf("parse response error: %w", err)
	}

	if w.config.Destination == "" {
		w.config.Destination = rspObj.Destination
	}
	w.config.RemotePort = rspObj.RemotePort

	cli.UpdateConnections(session.RemoteAddr().String(), rspObj.RemotePort, w.config.Destination, string(w.config.TunnelType), session.IsClosed())

	clis.ManagerTransport.PutConfig(w.config)
	w.multiClient.sessions.Store(w.config.ProxyId, session)
	log.Info("Open %v tunnel client success:%v:%v", w.config.TunnelType, w.config.ProxyId, rspObj.RemotePort)
	//Only http client open session.
	w.onlyOpenHttp(req.ProxyId)
	return nil
}

func (m *MultipleTunnelClient) Close() {
	m.closeOnce.Do(func() {
		// Close all sessions
		m.sessions.Range(func(key string, session *smux.Session) bool {
			if session != nil && !session.IsClosed() {
				_ = session.Close()
			}
			return true
		})
		m.sessions.Clear()
	})
}

func (w *tunnelClientWrapper) Close() {
	// Wrapper doesn't close the global client
	// Only remove this session if needed
	if w.config != nil {
		if session, ok := w.multiClient.sessions.Load(w.config.ProxyId); ok {
			if session != nil && !session.IsClosed() {
				_ = session.Close()
			}
			w.multiClient.sessions.Delete(w.config.ProxyId)
		}
	}
}

func (w *tunnelClientWrapper) onlyOpenHttp(proxyId string) {
	if w.config.TunnelType != lang.Http && w.config.TunnelType != lang.Https {
		return
	}
	req := &exchange.WorkConnReq{
		ProxyId: proxyId,
	}
	request, err := exchange.NewRequest(req)
	if err != nil {
		log.Error("Create WorkConnReq error: %v", err)
		return
	}
	for i := 0; i < httpWorkerConnCount; i++ {
		if err := clis.ManagerTransport.PushMessage(request); err != nil {
			log.Error("Push WorkConnReq message error: %v", err)
		}
	}
}
