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
	"time"

	"github.com/brook/client/cli"
	"github.com/brook/client/clis"
	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	"github.com/xtaci/smux"
)

// This function is used to register a new tunnel client for the TCP protocol
func init() {
	client := &MultipleTunnelClient{
		messageState: true,
		sessions:     hash.NewSyncMap[string, *smux.Session](),
	}
	ft := func(config *configs.ClientTunnelConfig) clis.TunnelClient {
		client.currentConfig = config
		if client.messageState {
			client.messageLister()
			client.messageState = false
		}
		return client
	}
	clis.RegisterTunnelClient(lang.Tcp, ft)
	clis.RegisterTunnelClient(lang.Udp, ft)
	clis.RegisterTunnelClient(lang.Http, ft)
	clis.RegisterTunnelClient(lang.Https, ft)
}

type MultipleTunnelClient struct {
	currentConfig *configs.ClientTunnelConfig
	messageState  bool
	sessions      *hash.SyncMap[string, *smux.Session]
}

func (m *MultipleTunnelClient) Done() <-chan struct{} {
	panic("implement me")
}

func (m *MultipleTunnelClient) GetName() string {
	return "multiple-tunnel-client"
}

func (m *MultipleTunnelClient) messageLister() {
	clis.ManagerTransport.AddMessageNotify(exchange.WorkerConnReq, func(r *exchange.Protocol) error {
		threading.GoSafe(func() {
			reqWorder, _ := exchange.Parse[exchange.WorkConnReq](r.Data)
			config := clis.ManagerTransport.GetConfig(reqWorder.ProxyId)
			if config == nil {
				log.Warn("configs is nil %v", reqWorder.ProxyId)
				return
			}
			id := config.ProxyId
			session, b := m.sessions.Load(id)
			if !b {
				log.Warn("not found session %v", reqWorder.ProxyId)
				return
			}
			config.RemotePort = reqWorder.RemotePort
			client := newTunnelClient(config, m)
			if client != nil {
				_ = client.Open(session)
			}
		})
		return nil
	})
}

func newTunnelClient(config *configs.ClientTunnelConfig, m *MultipleTunnelClient) clis.TunnelClient {
	switch config.TunnelType {
	case lang.Tcp:
		return NewTcpTunnelClient(config, m)
	case lang.Udp:
		return NewUdpTunnelClient(config, m)
	case lang.Http, lang.Https:
		return NewHttpTunnelClient(config)
	}
	return nil
}

// Open This function opens a TCP tunnel server for a given session.
func (m *MultipleTunnelClient) Open(session *smux.Session) error {
	// Create a new OpenTunnelReq struct with the proxy ID, tunnel type, and tunnel port.
	req := &exchange.OpenTunnelReq{
		ProxyId: m.currentConfig.ProxyId,
		UnId:    clis.ManagerTransport.UnId,
	}
	rsp, err := clis.ManagerTransport.SyncWrite(req, 5*time.Second)
	if err != nil {
		log.Error("Open %v tunnel server error %v", m.currentConfig.ProxyId, err)
		return err
	}
	if !rsp.IsSuccess() {
		log.Error("Open %v tunnel server error %v", m.currentConfig.ProxyId, rsp.RspMsg)
		return err
	}
	rspObj, _ := exchange.Parse[exchange.OpenTunnelResp](rsp.Data)

	cli.UpdateConnections(session.RemoteAddr().String(), rspObj.RemotePort, m.currentConfig.Destination, string(m.currentConfig.TunnelType), session.IsClosed())

	clis.ManagerTransport.PutConfig(m.currentConfig)
	m.sessions.Store(m.currentConfig.ProxyId, session)
	log.Info("Open %v tunnel client success:%v:%v", m.currentConfig.TunnelType, m.currentConfig.ProxyId, rspObj.RemotePort)
	//Only httpx client open session.
	m.OnlyOpenHttp(req.ProxyId, rspObj.RemotePort)
	return nil
}

func (m *MultipleTunnelClient) Close() {
	m.sessions.Clear()
	m.currentConfig = nil

}

func (m *MultipleTunnelClient) OnlyOpenHttp(proxyId string, remotePort int) {
	if m.currentConfig.TunnelType != lang.Http && m.currentConfig.TunnelType != lang.Https {
		return
	}
	req := &exchange.WorkConnReq{
		ProxyId:    proxyId,
		RemotePort: remotePort,
	}
	request, _ := exchange.NewRequest(req)
	for i := 0; i < 2; i++ {
		_ = clis.ManagerTransport.PushMessage(request)
	}
}
