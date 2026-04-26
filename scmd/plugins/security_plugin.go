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
	"sync/atomic"
	"time"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/lang"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/modules"
	"github.com/g-brook/brook/common/threading"
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
	security atomic.Value // *ipSecuritySnapshot
}

func (b *SecurityPlugin) Bind(cfg *configs.ServerTunnelConfig) {
	b.refreshIpSecurity(cfg)
	threading.GoSafe(func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				b.refreshIpSecurity(cfg)
			}
		}
	})
}

func (b *SecurityPlugin) refreshIpSecurity(cfg *configs.ServerTunnelConfig) {
	if cfg == nil || cfg.IpStrategy == "" {
		b.security.Store(&ipSecuritySnapshot{})
		return
	}
	security, err := service.SelectIpSecurity(cfg.IpStrategy)
	if err != nil || security == nil {
		b.security.Store(&ipSecuritySnapshot{})
		return
	}

	ips := make([]string, 0, len(security.Ips))
	for _, ip := range security.Ips {
		if ip == "" {
			continue
		}
		ips = append(ips, ip)
	}

	b.security.Store(&ipSecuritySnapshot{
		ips:          ips,
		strategy:     security.Strategy,
		strategyName: security.Name,
	})
}
func (b *SecurityPlugin) Open(ch trp.Channel, traverse srv.TraverseBy) error {
	s := b.getSecuritySnapshot()
	if s == nil || len(s.ips) == 0 || s.strategy == "" {
		traverse()
		return nil
	}
	ipStr, err := remoteIP(ch)
	if err != nil {
		log.Warn("security_plugin: parse remote ip failed: %v", err)
		traverse()
		return nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Warn("%s:invalid ip,reject.", ipStr)
		return errors.New("invalid ip address")
	}
	if s.strategy == lang.StrategyIntranet {
		if !isPrivateOrLoopbackIP(ip) {
			log.Warn("%s:non-private ip,reject.", ipStr)
			return errors.New("invalid ip address")
		}
	}
	match, _ := matchIPCIDR(ipStr, s.ips)
	switch s.strategy {
	case lang.StrategyWhite, lang.StrategyIntranet:
		if !match {
			log.Warn("%s:WhiteList or (IntranetList) name:%s,reject.", ipStr, s.strategyName)
			return errors.New("invalid ip address")
		}
	case lang.StrategyBlack:
		if match {
			log.Warn("%s:BlackList name:%s,reject.", ipStr, s.strategyName)
			return errors.New("invalid ip address")
		}
	}
	traverse()
	return nil
}

type ipSecuritySnapshot struct {
	ips          []string
	strategy     string
	strategyName string
}

func (b *SecurityPlugin) getSecuritySnapshot() *ipSecuritySnapshot {
	v := b.security.Load()
	if v == nil {
		return nil
	}
	s, _ := v.(*ipSecuritySnapshot)
	return s
}

func remoteIP(ch trp.Channel) (string, error) {
	if ch == nil {
		return "", errors.New("channel is nil")
	}
	addr := ch.RemoteAddr()
	if addr == nil {
		return "", errors.New("remote addr is nil")
	}
	switch a := addr.(type) {
	case *net.TCPAddr:
		if a.IP == nil {
			return "", errors.New("tcp remote ip is nil")
		}
		return a.IP.String(), nil
	case *net.UDPAddr:
		if a.IP == nil {
			return "", errors.New("udp remote ip is nil")
		}
		return a.IP.String(), nil
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return "", err
		}
		return host, nil
	}
}

func isPrivateOrLoopbackIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	return ip.IsPrivate() || ip.IsLoopback()
}

func matchIPCIDR(ipStr string, cidrList []string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, errors.New("invalid ip")
	}
	validRules := 0
	for _, cidr := range cidrList {
		switch cidr {
		case "0.0.0.0/0", "::/0":
			return true, nil
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		validRules++
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	if validRules == 0 && len(cidrList) > 0 {
		return false, errors.New("invalid cidr rules")
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
