package base

import (
	"github.com/brook/common/configs"
)

type ConfigNode struct {
	config *configs.ServerTunnelConfig
	// state is start.
	state bool
}

type TunnelConfigApi interface {
	// GetConfig Get tunnel config by proxy id
	GetConfig(proxyId string) *ConfigNode
}
