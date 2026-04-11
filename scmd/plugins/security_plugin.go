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

	"github.com/g-brook/brook/common/modules"
	trp "github.com/g-brook/brook/common/transport"
	"github.com/g-brook/brook/server/srv"
)

var securityModeName = modules.ModuleID("security_plugin")

func init() {
	modules.RegisterModule(&SecurityPlugin{})
}

type SecurityPlugin struct {
	srv.BaseServerHandler
}

func (b *SecurityPlugin) Open(ch trp.Channel, traverse srv.TraverseBy) error {
	addr := ch.RemoteAddr()
	whiteList := []string{
		"203.0.113.0/24",
		"113.111.1.0/24",
		"127.0.0.0/24",
	}
	ipAddr := addr.(*net.TCPAddr)
	ipcidr, err := matchIPCIDR(ipAddr.IP.String(), whiteList)
	if !ipcidr || err != nil {
		return err
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

func (s *SecurityPlugin) Module() modules.ModuleInfo {
	return modules.ModuleInfo{
		ID:         securityModeName,
		ModuleType: modules.TunnelPluginsModule,
		New: func() modules.Module {
			return new(SecurityPlugin)
		},
	}
}
