package run

import (
	_ "github.com/brook/client/clients"
	"github.com/brook/common/log"
	"github.com/brook/common/srv"
)

func LoadTunnel() {
	clients := srv.GetTunnelClients()
	for k, _ := range clients {
		log.Info("Loading tunnel  %v client,", k)
	}
}
