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

package plugins

import (
	"errors"
	"net"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/modules"
	trp "github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/scmd/web/service"
	"github.com/g-brook/brook/server/srv"
)

var securityModeName = modules.ModuleID("security_plugin")

func init() {
	modules.RegisterModule(&SecurityPlugin{})
}

type SecurityPlugin struct {
	srv.BaseServerHandler
	ips      []string
	strategy string
}

func (b *SecurityPlugin) Bind(cfg *configs.ServerTunnelConfig) {
	security, err := service.SelectIpSecurity(cfg.IpStrategy)
	if err == nil && security != nil {
		b.ips = security.Ips
		b.strategy = security.Strategy
	}
}

func (b *SecurityPlugin) Open(ch trp.Channel, traverse srv.TraverseBy) error {
	addr := ch.RemoteAddr()
	ipAddr := addr.(*net.TCPAddr)
	match, err := matchIPCIDR(ipAddr.IP.String(), b.ips)
	switch b.strategy {
	case "WL":
		if !match || err != nil {
			return errors.New("invalid ip address")
		}
		break
	case "BL":
		if match || err != nil {
			return errors.New("invalid ip address")
		}
		break
	}
	traverse()
	return nil
}

func matchIPCIDR(ipStr string, cidrList []string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, errors.New("invalid ip")
	}
	for _, cidr := range cidrList {
		switch cidr {
		case "0.0.0.0/0", "::/0", "0.0.0.0/24":
			return true, nil
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

func (b *SecurityPlugin) Module() modules.ModuleInfo {
	return modules.ModuleInfo{
		ID:         securityModeName,
		ModuleType: modules.TunnelPluginsModule,
		New: func() modules.Module {
			return new(SecurityPlugin)
		},
	}
}
