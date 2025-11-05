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
	"sync"

	"github.com/brook/common/configs"
	"github.com/brook/common/hash"
	"github.com/brook/common/lang"
	"github.com/brook/common/queue"
	"github.com/brook/common/threading"
)

type ConfigNode struct {
	config *configs.ServerTunnelConfig
	// state is start.
	state    bool
	openLock sync.Mutex
}

type ConfigNotify func(cfg *ConfigNode)

type TunnelConfigApi interface {

	// GetConfig Get tunnel configs by proxy id
	GetConfig(proxyId string) *ConfigNode

	UpdateConfig(proxyId string) *ConfigNode
}

var TunnelCfm = &ConfigManager{
	queue:   queue.NewMemoryQueue[string](100),
	listens: hash.NewSyncMap[string, ConfigNotify](),
}

func TransformProtocol(protocol string) lang.TunnelType {
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

type ConfigManager struct {
	ConfigApi TunnelConfigApi
	queue     *queue.MemoryQueue[string]
	listens   *hash.SyncMap[string, ConfigNotify]
}

func (receiver *ConfigManager) AddListen(proxyId string, notify ConfigNotify) {
	receiver.listens.Store(proxyId, notify)
}

func (receiver *ConfigManager) Push(proxyId string) {
	receiver.queue.Push(proxyId)
}

func (receiver *ConfigManager) Running(api TunnelConfigApi) {
	receiver.ConfigApi = api
	threading.GoSafe(func() {
		for {
			proxyId := receiver.queue.Pop()
			if proxyId != "" {
				newConfig := receiver.ConfigApi.UpdateConfig(proxyId)
				if newConfig != nil {
					load, b := receiver.listens.Load(newConfig.config.Id)
					if b {
						load(newConfig)
					}
				}
			}
		}
	})
}
