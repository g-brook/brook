package base

import (
	"github.com/brook/common/configs"
	"github.com/brook/common/hash"
	"github.com/brook/common/queue"
)

type ConfigNode struct {
	config *configs.ServerTunnelConfig
	// state is start.
	state bool
}

type ConfigNotify func(cfg *ConfigNode)

type TunnelConfigApi interface {

	// GetConfig Get tunnel config by proxy id
	GetConfig(proxyId string) *ConfigNode

	UpdateConfig(proxyId string) *ConfigNode
}

var CFM = &ConfigManager{
	queue:   queue.NewMemoryQueue[string](100),
	listens: hash.NewSyncMap[string, ConfigNotify](),
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
	go func() {
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
	}()
}
