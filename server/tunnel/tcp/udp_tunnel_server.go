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

package tcp

import (
	"errors"
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/log"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

type TunnelUdpServer struct {
	*tunnel.BaseTunnelServer
	registerLock sync.Mutex
	resources    *Resources
}

// NewUdpTunnelServer creates a new TCP tunnel server instance
func NewUdpTunnelServer(server *tunnel.BaseTunnelServer) *TunnelUdpServer {
	tunnelServer := &TunnelUdpServer{
		BaseTunnelServer: server,
		resources:        NewResources(100, server.Cfg, server.GetManager),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *TunnelUdpServer) RegisterConn(ch trp.Channel, request exchange.TRegister) (serverId string, err error) {
	if request.GetProxyId() == "" {
		log.Warn("Register udp tunnel, but It' proxyId is nil")
		return "", errors.New("it' proxyId is nil")
	}
	if _, ok := ch.(*trp.SChannel); ok {
		htl.registerLock.Lock()
		defer htl.registerLock.Unlock()
		serverId, err = htl.BaseTunnelServer.RegisterConn(ch, request)
		log.Info("Register udp tunnel, proxyId: %s", request.GetProxyId())
		return
	}
	return "", errors.New("unknown channel type")
}

func (htl *TunnelUdpServer) OpenWorker(ch trp.Channel, request *exchange.ClientWorkConnReq) error {
	id := request.ServerId
	ch, b := htl.TunnelChannel.Load(id)
	if b && !ch.IsClose() {
		_ = htl.resources.put(ch)
		log.Info("dup add user connection, proxyId: %s", request.ProxyId)
		return nil
	}
	return errors.New("channel is nil or closed")
}

func (htl *TunnelUdpServer) Reader(ch trp.Channel, tb srv.TraverseBy) error {
	switch workConn := ch.(type) {
	case srv.GContext:
		userConn, _ := htl.resources.get()
		if userConn == nil {
			_ = ch.Close()
			return nil
		}
		data, _ := workConn.Next(-1)
		userConn.(*UdpSChannel).AsyncWriter(data, ch)
		_ = htl.resources.put(userConn)
		return nil
	}
	tb()
	return nil
}

func (htl *TunnelUdpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("udp tunnel server started:%v", htl.Port())
	return nil
}
