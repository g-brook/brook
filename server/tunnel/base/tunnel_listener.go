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

package base

import (
	"fmt"
	"sync"
	"time"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/exchange"
	"github.com/g-brook/brook/common/hash"
	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/log"
	. "github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/remote"
	"github.com/g-brook/brook/server/srv"
	"github.com/g-brook/brook/server/tunnel"
	"github.com/g-brook/brook/server/tunnel/http"
	"github.com/g-brook/brook/server/tunnel/tcp"
)

var servers *hash.SyncMap[string, tunnel.TunnelServer]

func init() {
	servers = hash.NewSyncMap[string, tunnel.TunnelServer]()
	remote.OpenTunnelServerFun = OpenTunnelServer
	remote.ManagerTunnelServerFun = ManagerTunnelServe
	remote.RegisterVisitorFun = RegisterVisitor
}

func RegisterVisitor(req *exchange.VisitorRegister, ch Channel) error {
	cfgNode := TunnelCfm.ConfigApi.GetConfig(req.ProxyId)
	if cfgNode == nil {
		return fmt.Errorf("visitor config not found proxy id %v", req.ProxyId)
	}
	cfgNode.openLock.Lock()
	defer cfgNode.openLock.Unlock()
	server, b := servers.Load(cfgNode.Config.Id)
	if !b {
		return fmt.Errorf("visitor server not start proxy id %v", req.ProxyId)
	}
	if server.GetConfig().Visitor == nil {
		return fmt.Errorf("visitor config not found proxy id %v", req.ProxyId)
	}
	if server.GetConfig().Visitor.Token != req.Token {
		return fmt.Errorf("visitor token not match %v", req.ProxyId)
	}
	visitorServer, ok := server.GetServer().(*srv.VisitorServer)
	if !ok || visitorServer == nil {
		return fmt.Errorf("server is not visitor server proxy id %v", req.ProxyId)
	}
	if err := visitorServer.AddLastChannel(ch); err != nil {
		return err
	}
	return nil
}

// OpenTunnelServer open tcp tunnel server
// This function opens a tunnel server based on the request parameters.
func OpenTunnelServer(request *exchange.OpenTunnelReq, manager Channel) (*remote.TunnelCfg, error) {
	cfgNode := TunnelCfm.ConfigApi.GetConfig(request.ProxyId)
	if cfgNode == nil {
		return nil, fmt.Errorf("not found proxy id %v", request.ProxyId)
	}
	cfgNode.openLock.Lock()
	defer cfgNode.openLock.Unlock()
	t, b := servers.Load(cfgNode.Config.Id)
	if b {
		t.PutManager(manager)
		return remote.NewTunnelCfg(cfgNode.Config.Port, cfgNode.Config.Destination), nil
	}
	baseServer, err := running(cfgNode.Config)
	if err != nil {
		return nil, err
	}
	t, b = servers.Load(cfgNode.Config.Id)
	if b {
		TunnelCfm.AddListen(cfgNode.Config.Id, func(cfg *ConfigNode) {
			baseServer.UpdateConfig(cfg.Config)
		})
		t.PutManager(manager)
	}
	return remote.NewTunnelCfg(baseServer.Port(), baseServer.Cfg.Destination), err
}

func ManagerTunnelServe(proxyId string, manager Channel) error {
	cfgNode := TunnelCfm.ConfigApi.GetConfig(proxyId)
	if cfgNode == nil {
		return fmt.Errorf("not found proxy id %v", proxyId)
	}
	cfgNode.openLock.Lock()
	defer cfgNode.openLock.Unlock()
	t, b := servers.Load(cfgNode.Config.Id)
	if b {
		t.PutManager(manager)
	}
	return nil
}

func running(config *configs.ServerTunnelConfig) (*tunnel.BaseTunnelServer, error) {
	isVisitor := config.ModelId == srv.VisitorServerPlugin
	baseServer := tunnel.NewBaseTunnelServer(config, config.ModelId)
	var server tunnel.TunnelServer
	var netWork lang.Network
	var err error
	if isVisitor {
		netWork, server, err = newVisitorTunnel(config, baseServer)
	} else {
		netWork, server = newNormalTunnel(config, baseServer)
	}
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, fmt.Errorf("the server %v:%s is not supported", config.Type, config.Id)
	}
	//Start the server.
	err = server.Start(netWork)
	if err != nil {
		//Release the port if the server fails to start.
		return nil, err
	}
	servers.Store(config.Id, server)
	return baseServer, nil
}

func newNormalTunnel(config *configs.ServerTunnelConfig, server *tunnel.BaseTunnelServer) (network lang.Network, tserver tunnel.TunnelServer) {
	switch config.Type {
	case lang.Http, lang.Https:
		tunnelServer, err := http.NewHttpTunnelServer(server)
		if err != nil {
			err := fmt.Errorf("the server %v:%s init error", config.Type, config.Id)
			log.Warn(err.Error())
			return "", nil
		}
		tserver = tunnelServer
		network = lang.NetworkTcp
		break
	case lang.Tcp:
		tserver = tcp.NewTcpTunnelServer(server)
		network = lang.NetworkTcp
		break
	case lang.Udp:
		tserver = tcp.NewUdpTunnelServer(server)
		network = lang.NetworkUdp
		break
	}
	return
}

func newVisitorTunnel(config *configs.ServerTunnelConfig, server *tunnel.BaseTunnelServer) (network lang.Network, tserver tunnel.TunnelServer, err error) {
	switch config.Type {
	case lang.Tcp:
		return lang.NetworkTcp, tcp.NewVTcpTunnelServer(server), nil
	case lang.Udp:
		return lang.NetworkUdp, tcp.NewUdpTunnelServer(server), nil
	case lang.Http, lang.Https:
		tunnelServer, err := http.NewHttpTunnelServer(server)
		if err != nil {
			return "", nil, fmt.Errorf("the visitor server %v:%s init error: %w", config.Type, config.Id, err)
		}
		return lang.NetworkTcp, tunnelServer, nil
	}
	return "", nil, fmt.Errorf("visitor tunnel type %s is not supported", config.Type)
}

type PortPool struct {
	mu      sync.Mutex
	ports   map[int]time.Time
	ttl     time.Duration
	minPort int
	maxPort int
}
