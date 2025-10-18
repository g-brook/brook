package base

import (
	"errors"

	sf "github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/queue"
	"github.com/brook/common/utils"
	"github.com/brook/server/tunnel/http"
	"github.com/brook/server/tunnel/tcp"
)

var ServerQueue = queue.NewMemoryQueue[*sf.ServerTunnelConfig](100)

func init() {
	lister()
}

func RunTunnelServer(config *sf.ServerConfig) {
	tunnelConfig := GetTunnelConfig(config)
	for _, tcf := range tunnelConfig {
		runServer(tcf)
	}
}

func RunServer(proxyId string) error {
	proxy := GetTunnelConfigByProxy(proxyId)
	if proxy == nil {
		log.Error("proxyId not found: %v", proxy)
		return errors.New("proxyId not found")
	}
	ServerQueue.Push(proxy)
	return nil
}

func lister() {
	go func() {
		for {
			config := ServerQueue.Pop()
			runServer(config)
		}
	}()
}

func runServer(config *sf.ServerTunnelConfig) {
	if config != nil {
		switch config.Type {
		case utils.Http, utils.Https:
			http.RunStart(config)
			break
		case utils.Tcp, utils.Udp:
			tcp.RunStart(config)
			break
		}
	}
}
