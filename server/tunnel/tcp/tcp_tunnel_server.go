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

	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/iox"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	trp "github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/defin"
	"github.com/g-brook/brook/server/srv"
	"github.com/g-brook/brook/server/tunnel"
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
		resources:        NewResources(100, server.Cfg, server.GetManager),
	}
	server.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func (htl *TunnelTcpServer) RegisterConn(ch trp.Channel, request exchange.TRegister) (serverId string, err error) {
	if request.GetProxyId() == "" {
		log.Warn("Register tcp tunnel, but It' proxyId is nil")
		return "", errors.New("it' proxyId is nil")
	}
	htl.registerLock.Lock()
	defer htl.registerLock.Unlock()
	serverId, err = htl.BaseTunnelServer.RegisterConn(ch, request)
	log.Info("Register tcp tunnel, proxyId: %s", request.GetProxyId())
	return
}

func (htl *TunnelTcpServer) OpenWorker(ch trp.Channel, request *exchange.ClientWorkConnReq) error {
	// Open a new goroutine to handle the channel
	id := request.ServerId
	ch, b := htl.TunnelChannel.Load(id)
	if b && !ch.IsClose() {
		_ = htl.resources.put(ch)
		log.Info("add user connection, proxyId: %s", request.ProxyId)
		return nil
	}
	return errors.New("channel is nil or closed")
}

func (htl *TunnelTcpServer) Reader(ch trp.Channel, _ srv.TraverseBy) error {
	switch workConn := ch.(type) {
	case srv.GContext:
		chId, ok := workConn.GetContext().GetAttr(defin.ToSChannelId)
		if ok && chId != "" {
			dest, ok := htl.TunnelChannel.Load(chId.(string))
			if ok {
				srcBytes, err := workConn.Next(-1)
				if err != nil {
					log.Debug("iox.copy error %v", err)
					return err
				}
				_, err = dest.Write(srcBytes)
			}
		}
	}
	return nil
}

func (htl *TunnelTcpServer) Open(ch trp.Channel, _ srv.TraverseBy) error {
	userConn, err := htl.resources.get()
	if userConn == nil || err != nil {
		_ = ch.Close()
		return err
	}
	switch workConn := ch.(type) {
	case srv.GContext:
		workConn.GetContext().AddAttr(defin.ToSChannelId, userConn.GetId())
		threading.GoSafe(func() {
			err := iox.SinglePipe(userConn, workConn.(trp.Channel))
			log.Debug("iox.SinglePipe error %v", err)
		})
	}
	return err
}

func (htl *TunnelTcpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("TCP tunnel server started:%v", htl.Port())
	return nil
}
