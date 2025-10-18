package base

import (
	"encoding/json"

	sf "github.com/brook/common/configs"
	"github.com/brook/common/log"
	"github.com/brook/common/utils"
	"github.com/brook/server/web/sql"
)

// GetTunnelConfig retrieves the server tunnel configuration
// This function is used to obtain the configuration settings for establishing a server tunnel
// It returns a ServerTunnelConfig struct which contains all necessary parameters for tunnel setup
func GetTunnelConfig(sc *sf.ServerConfig) []*sf.ServerTunnelConfig {
	if !sc.EnableWeb {
		return sc.Tunnel
	}
	config := sql.GetAllProxyConfig()
	var list []*sf.ServerTunnelConfig
	for _, item := range config {
		tunnelConfig := format(item)
		if tunnelConfig != nil {
			list = append(list, tunnelConfig)
		}
	}
	return list
}

func GetTunnelConfigByProxy(proxyId string) *sf.ServerTunnelConfig {
	proxyConfig := sql.GetAllProxyConfigByProxyId(proxyId)
	if proxyConfig == nil {
		return nil
	}
	return format(proxyConfig)

}

func getTunnelWebConfig(stc *sf.ServerTunnelConfig) bool {
	config := sql.GetWebProxyConfig(stc.Id)
	if config != nil {
		proxy := config.Proxy
		stc.KeyFile = config.KeyFile
		stc.CertFile = config.CertFile
		_ = json.Unmarshal([]byte(proxy), &stc.Proxy)
		return true
	}
	return false
}

func format(item *sql.ProxyConfig) *sf.ServerTunnelConfig {
	var st = new(sf.ServerTunnelConfig)
	st.Id = item.ProxyID
	st.Port = item.RemotePort
	protocol := transformProtocol(item.Protocol)
	if protocol == "" {
		log.Error("protocol is not support: %s", item.Protocol)
	} else {
		st.Type = protocol
		if st.Type == utils.Http || st.Type == utils.Https {
			if getTunnelWebConfig(st) {
				return st
			}
		} else {
			return st
		}
	}
	return nil
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
