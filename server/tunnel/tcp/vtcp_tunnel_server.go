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
	"fmt"
	"sync"

	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/log"
	trp "github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/srv"
	"github.com/g-brook/brook/server/tunnel"
)

type TunnelVTcpServer struct {
	*tunnel.BaseTunnelServer

	registerLock sync.Mutex

	resources *Resources
}

func NewVTcpTunnelServer(server *tunnel.BaseTunnelServer) *TunnelVTcpServer {
	tunnelServer := &TunnelVTcpServer{
		BaseTunnelServer: server,
		resources:        NewResources(100, server.Cfg, server.GetManager),
	}
	tunnelServer.DoStart = tunnelServer.startAfter
	return tunnelServer
}

func notifyVisitorOpenReady(ch trp.Channel, err error) {
	ready, ok := ch.GetAttr(srv.VisitorOpenReadyKey)
	if !ok {
		return
	}
	if notify, ok := ready.(srv.VisitorOpenReadyFunc); ok {
		notify(err)
	}
}

func (htl *TunnelVTcpServer) Reader(ch trp.Channel, _ srv.TraverseBy) error {
	fmt.Println("xxxdddddddd")
	if context, ok := ch.(*srv.VisitorContextChannel); ok {
		context.Pipe()
		return srv.SkipError
	}
	return errors.New("not a visitor channel")
}

func (htl *TunnelVTcpServer) Open(ch trp.Channel, _ srv.TraverseBy) error {
	conn, err := htl.resources.get()
	if err != nil {
		notifyVisitorOpenReady(ch, err)
		return err
	}
	if context, ok := ch.(*srv.VisitorContextChannel); ok {
		context.GetVisitorChannel().TargetChannel(conn)
		notifyVisitorOpenReady(ch, nil)
		return nil
	}
	return errors.New("not supported")
}

func (htl *TunnelVTcpServer) OpenWorker(ch trp.Channel, request *exchange.ClientWorkConnReq) error {
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

func (htl *TunnelVTcpServer) startAfter() error {
	tunnel.AddTunnel(htl)
	htl.Server.AddHandler(htl)
	log.Info("VTCP tunnel server started:%v", htl.Port())
	return nil
}
