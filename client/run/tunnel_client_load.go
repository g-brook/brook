package run

import (
	"github.com/brook/client/clis"
	_ "github.com/brook/client/tunnel"
	"github.com/brook/common/log"
)

func LoadTunnel() {
	clients := clis.GetTunnelClients()
	for k, _ := range clients {
		log.Info("Loading tunnel  %v client,", k)
	}
}
