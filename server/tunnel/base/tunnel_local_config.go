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
	"encoding/json"

	sf "github.com/brook/common/configs"
	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/server/web/sql"
)

type LocalTunnelConfig struct {
	configs *hash.SyncMap[string, *ConfigNode]
}

// GetConfig retrieves a server tunnel configuration by proxy ID
//
// Parameters:
//   - proxyId: string identifier of the proxy to retrieve configuration for
//
// Returns:
//   - *sf.ServerTunnelConfig: pointer to the server tunnel configuration associated with the proxy ID
func (receiver *LocalTunnelConfig) GetConfig(proxyId string) *ConfigNode {
	load, _ := receiver.configs.Load(proxyId) // Load retrieves the value for a key from the sync.Map
	return load
}

func (receiver *LocalTunnelConfig) UpdateConfig(proxyId string) *ConfigNode {
	cfg := sql.GetProxyConfigByProxyId(proxyId)
	if cfg == nil {
		return nil
	}
	config := format(cfg)
	if config != nil {
		receiver.configs.Store(proxyId, &ConfigNode{
			config: config,
			state:  false,
		})
	}
	return receiver.GetConfig(proxyId)
}

func NewLocalTunnelConfig() *LocalTunnelConfig {
	return &LocalTunnelConfig{
		configs: hash.NewSyncMap[string, *ConfigNode](),
	}
}

func InitTunnelConfig(sc *sf.ServerConfig) {
	var ltc = NewLocalTunnelConfig()
	for _, item := range getTunnelConfig(sc) {
		ltc.configs.Store(item.Id, &ConfigNode{
			config: item,
			state:  false,
		})
	}
	CFM.Running(ltc)
}

// GetTunnelConfig retrieves the server tunnel configuration
// This function is used to obtain the configuration settings for establishing a server tunnel
// It returns a ServerTunnelConfig struct which contains all necessary parameters for tunnel setup
func getTunnelConfig(sc *sf.ServerConfig) []*sf.ServerTunnelConfig {
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

func getTunnelWebConfig(stc *sf.ServerTunnelConfig, refProxyId int) bool {
	config := sql.GetWebProxyConfig(refProxyId)
	if config != nil {
		proxy := config.Proxy
		stc.KeyFile = config.KeyFile
		stc.CertFile = config.CertFile
		_ = json.Unmarshal([]byte(proxy), &stc.Http)
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
		if st.Type == lang.Http || st.Type == lang.Https {
			if getTunnelWebConfig(st, item.Idx) {
				return st
			}
		} else {
			return st
		}
	}
	return nil
}

func transformProtocol(protocol string) lang.TunnelType {
	switch protocol {
	case "HTTP":
		return lang.Http
	case "HTTPS":
		return lang.Https
	case "TCP":
		return lang.Tcp
	case "UDP":
		return lang.Udp
	default:
		return ""
	}
}
