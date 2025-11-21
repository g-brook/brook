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

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	. "github.com/brook/common/transport"
	"github.com/brook/server/remote"
	"github.com/brook/server/tunnel"
	"github.com/brook/server/tunnel/http"
	"github.com/brook/server/tunnel/tcp"
)

var servers *hash.SyncMap[string, tunnel.TunnelServer]

func init() {
	servers = hash.NewSyncMap[string, tunnel.TunnelServer]()
	remote.OpenTunnelServerFun = OpenTunnelServer
}

// OpenTunnelServer open tcp tunnel server
// This function opens a tunnel server based on the request parameters.
func OpenTunnelServer(request exchange.OpenTunnelReq, manager Channel) (*remote.TunnelCfg, error) {
	cfgNode := TunnelCfm.ConfigApi.GetConfig(request.ProxyId)
	if cfgNode == nil {
		return nil, fmt.Errorf("not found proxy id %v", request.ProxyId)
	}
	cfgNode.openLock.Lock()
	defer cfgNode.openLock.Unlock()
	t, b := servers.Load(cfgNode.config.Id)
	if b {
		t.PutManager(manager)
		return remote.NewTunnelCfg(cfgNode.config.Port, cfgNode.config.Destination), nil
	} else {
		baseServer, err := running(cfgNode.config)
		if err != nil {
			return nil, err
		}
		t, b := servers.Load(cfgNode.config.Id)
		if b {
			TunnelCfm.AddListen(cfgNode.config.Id, func(cfg *ConfigNode) {
				baseServer.UpdateConfig(cfg.config)
			})
			t.PutManager(manager)
		}
		return remote.NewTunnelCfg(baseServer.Port(), baseServer.Cfg.Destination), err
	}
}

func running(config *configs.ServerTunnelConfig) (*tunnel.BaseTunnelServer, error) {
	baseServer := tunnel.NewBaseTunnelServer(config)
	var server tunnel.TunnelServer
	var netWork lang.Network
	if config.Type == lang.Tcp {
		server = tcp.NewTcpTunnelServer(baseServer)
		netWork = lang.NetworkTcp
	} else if config.Type == lang.Udp {
		server = tcp.NewUdpTunnelServer(baseServer)
		netWork = lang.NetworkUdp
	} else if config.Type == lang.Https || config.Type == lang.Http {
		tunnelServer, err := http.NewHttpTunnelServer(baseServer)
		if err != nil {
			return nil, fmt.Errorf("the server %v:%s init error", config.Type, config.Id)
		}
		server = tunnelServer
		netWork = lang.NetworkTcp
	} else {
		return nil, fmt.Errorf("not support tunnel type %v", config.Type)
	}
	//Start the server.
	err := server.Start(netWork)
	if err != nil {
		//Release the port if the server fails to start.
		return nil, err
	}
	servers.Store(config.Id, server)
	return baseServer, nil
}

type PortPool struct {
	mu      sync.Mutex
	ports   map[int]time.Time
	ttl     time.Duration
	minPort int
	maxPort int
}
