package run

import (
	sf "github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/web/sql"
)

// GetTunnelConfig retrieves the server tunnel configuration
// This function is used to obtain the configuration settings for establishing a server tunnel
// It returns a ServerTunnelConfig struct which contains all necessary parameters for tunnel setup
func GetTunnelConfig(sc sf.ServerConfig) []*sf.ServerTunnelConfig {
	if !sc.EnableWeb {
		return sc.Tunnel
	}
	config := sql.GetAllProxyConfig()
	var list []*sf.ServerTunnelConfig
	for _, item := range config {
		var st = new(sf.ServerTunnelConfig)
		st.Id = item.ProxyID
		st.Port = item.RemotePort
		protocol := transformProtocol(item.Protocol)
		if protocol == "" {
			log.Error("protocol is not support: %s", item.Protocol)
		} else {
			st.Type = protocol
			list = append(list, st)
		}
	}
	return list
}

func transformProtocol(protocol string) utils.TunnelType {
	switch protocol {
	case "HTTP":
		return utils.Http
	case "HTTPS":
		return utils.Https
	case "TCP":
		return utils.Tcp
	case "UDP":
		return utils.Udp
	default:
		return ""
	}
}
