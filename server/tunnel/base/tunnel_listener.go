package base

import (
	"fmt"
	"sync"
	"time"

	"github.com/brook/common/configs"
	"github.com/brook/common/exchange"
	"github.com/brook/common/hash"
	. "github.com/brook/common/transport"
	"github.com/brook/common/utils"
	"github.com/brook/server/remote"
	"github.com/brook/server/tunnel"
	"github.com/brook/server/tunnel/http"
	"github.com/brook/server/tunnel/tcp"
)

var tunnelAPI TunnelConfigApi

var servers *hash.SyncMap[string, tunnel.TunnelServer]

func init() {
	servers = hash.NewSyncMap[string, tunnel.TunnelServer]()
	remote.OpenTunnelServerFun = OpenTunnelServer
}

// OpenTunnelServer open tcp tunnel server
// This function opens a tunnel server based on the request parameters.
func OpenTunnelServer(request exchange.OpenTunnelReq, manager Channel) (int, error) {
	cfgNode := tunnelAPI.GetConfig(request.ProxyId)
	if cfgNode == nil {
		return 0, fmt.Errorf("not found proxy id %v", request.ProxyId)
	}
	t, b := servers.Load(cfgNode.config.Id)
	if b {
		t.PutManager(manager)
		return cfgNode.config.Port, nil
	} else {
		i, err := running(cfgNode.config)
		t, b := servers.Load(cfgNode.config.Id)
		if b {
			t.PutManager(manager)
		}
		return i, err
	}
}

func running(config *configs.ServerTunnelConfig) (int, error) {
	baseServer := tunnel.NewBaseTunnelServer(config)
	var server tunnel.TunnelServer
	var netWork utils.Network
	if config.Type == utils.Tcp {
		server = tcp.NewTcpTunnelServer(baseServer)
		netWork = utils.NetworkTcp
	} else if config.Type == utils.Udp {
		server = tcp.NewUdpTunnelServer(baseServer)
		netWork = utils.NetworkUdp
	} else if config.Type == utils.Https || config.Type == utils.Http {
		server = http.NewHttpTunnelServer(baseServer)
		netWork = utils.NetworkTcp
	} else {
		return 0, fmt.Errorf("not support tunnel type %v", config.Type)
	}
	//Start the server.
	err := server.Start(netWork)
	if err != nil {
		//Release the port if the server fails to start.
		return 0, err
	}
	servers.Store(config.Id, server)
	return config.Port, nil
}

type PortPool struct {
	mu      sync.Mutex
	ports   map[int]time.Time
	ttl     time.Duration
	minPort int
	maxPort int
}
