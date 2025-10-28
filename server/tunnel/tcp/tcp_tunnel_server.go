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
	"sync"

	"github.com/brook/common/exchange"
	"github.com/brook/common/iox"
	"github.com/brook/common/log"
	"github.com/brook/common/threading"
	trp "github.com/brook/common/transport"
	"github.com/brook/server/defin"
	"github.com/brook/server/srv"
	"github.com/brook/server/tunnel"
)

type TunnelTcpServer struct {
	*tunnel.BaseTunnelServer
	registerLock sync.Mutex
	resources    *Resources
}

// NewTcpTunnelServer creates a new TCP tunnel server instance
func NewTcpTunnelServer(server *tunnel.BaseTunnelServer) *TunnelTcpServer {
	tunnelServer := &TunnelTcpServer{
		BaseTunnelServer: server,
		resources:        NewResources(100, server.Cfg.Id, server.Cfg.Port, server.GetManager),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *TunnelTcpServer) RegisterConn(ch trp.Channel, request exchange.TRegister) {
	if request.GetProxyId() == "" {
		log.Warn("Register tcp tunnel, but It' proxyId is nil")
		return
	}
	htl.registerLock.Lock()
	defer htl.registerLock.Unlock()
	htl.BaseTunnelServer.RegisterConn(ch, request)
	_ = htl.resources.put(ch)
	log.Info("Register tcp tunnel, proxyId: %s", request.GetProxyId())

}

func (htl *TunnelTcpServer) Reader(ch trp.Channel, _ srv.TraverseBy) {
	switch workConn := ch.(type) {
	case srv.GContext:
		chId, ok := workConn.GetContext().GetAttr(defin.ToSChannelId)
		if ok && chId != "" {
			dest, ok := htl.TunnelChannel[chId.(string)]
			if ok {
				err := iox.Copy(ch, dest)
				if err != nil {
					log.Debug("iox.copy error %v", err)
				}
			}
		}
	}
}

func (htl *TunnelTcpServer) Open(ch trp.Channel, _ srv.TraverseBy) {
	userConn, _ := htl.resources.get()
	if userConn == nil {
		_ = ch.Close()
		return
	}
	switch workConn := ch.(type) {
	case srv.GContext:
		workConn.GetContext().AddAttr(defin.ToSChannelId, userConn.GetId())
		threading.GoSafe(func() {
			err := iox.SinglePipe(userConn, workConn.(trp.Channel))
			log.Debug("iox.SinglePipe error %v", err)
		})
	}
}

func (htl *TunnelTcpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("TCP tunnel server started:%v", htl.Port())
	return nil
}
