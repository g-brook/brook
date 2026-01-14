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

package api

import (
	"encoding/json"

	"github.com/brook/common/configs"
	"github.com/brook/common/lang"
	"github.com/brook/common/log"
	"github.com/brook/server/defin"
	"github.com/brook/server/metrics"
	"github.com/brook/server/tunnel/base"
	"github.com/brook/server/web/errs"
	"github.com/brook/server/web/sql"
)

func init() {
	RegisterRoute(NewRoute("/getProxyConfigs", "POST"), getProxyConfigs)
	RegisterRoute(NewRoute("/addProxyConfigs", "POST"), addProxyConfigs)
	RegisterRoute(NewRoute("/delProxyConfigs", "POST"), delProxyConfig)
	RegisterRoute(NewRoute("/addWebConfigs", "POST"), addWebConfigs)
	RegisterRoute(NewRoute("/getWebConfigs", "POST"), getWebConfigs)
	RegisterRoute(NewRoute("/updateProxyConfig", "POST"), updateProxyConfig)
	RegisterRoute(NewRoute("/updateProxyState", "POST"), updateProxyState)
	RegisterRoute(NewRoute("/genClientConfig", "POST"), genClientConfig)
}

// getProxyConfigs retrieves configuration information from the database
// It takes a pointer to a Request with any type as parameter
// and returns a Response containing the configuration data or error information
func getProxyConfigs(*Request[any]) *Response {
	config := sql.QueryProxyConfig()
	if config == nil {
		return NewResponseSuccess(nil)
	}
	newConfig := make([]*ProxyConfig, len(config))
	for i, proxyConfig := range config {
		newConfig[i] = newProxyConfig(proxyConfig)
	}
	configMap := make(map[string]*ProxyConfig)
	for _, cf := range newConfig {
		configMap[cf.ProxyID] = cf
	}
	servers := metrics.M.GetServers()
	for _, server := range servers {
		proxyConfig, ok := configMap[server.Id()]
		if ok {
			proxyConfig.IsRunning = true
			proxyConfig.Runtime = server.Runtime().Format("2006-01-02 15:04:05")
			proxyConfig.Clients = server.Clients()
		}
	}
	for _, v := range configMap {
		if v.IsHttpOrHttps() {
			proxyConfig := sql.GetWebProxyConfig(v.Idx)
			v.IsExistWeb = proxyConfig != nil
		}
	}
	return NewResponseSuccess(newConfig)
}

func genClientConfig(*Request[any]) *Response {
	serverPort := defin.GetServerPort()
	cfgs := sql.GetAllProxyConfig()
	var tunnelCfgs = make([]*configs.ClientTunnelConfig, 0)
	if cfgs != nil {
		for _, cfg := range cfgs {
			s := cfg.Destination.String
			if s == "" {
				s = "#{localAddress}"
			}
			tcfg := &configs.ClientTunnelConfig{
				TunnelType:  base.TransformProtocol(cfg.Protocol),
				ProxyId:     cfg.ProxyID,
				Destination: s,
			}
			if tcfg.ProxyId == "UDP" {
				tcfg.UdpSize = 1500
			}
			if cfg.Protocol == "HTTP" || cfg.Protocol == "HTTPS" {
				webCfg, ok := getWebConfig(cfg.Idx)
				if ok {
					for _, proxyInfo := range webCfg.Proxy {
						tcfg.HttpId = proxyInfo.Id
					}
				}
			}
			tunnelCfgs = append(tunnelCfgs, tcfg)
		}
	}
	cfg := &configs.ClientConfig{
		ServerHost: "#{host}",
		ServerPort: serverPort,
		Tunnels:    tunnelCfgs,
		PingTime:   lang.DefaultPingTime,
		Token:      defin.GetToken(),
		Logger:     nil,
	}
	return NewResponseSuccess(cfg)
}

func delProxyConfig(req *Request[ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.DelProxyConfig(req.Body.Idx)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete proxy configs failed")
	}
	toPushConfig(req.Body.Idx)
	return NewResponseSuccess(nil)
}

func updateProxyConfig(req *Request[ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.UpdateProxyConfig(req.Body.toDb())
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "update proxy configs failed")
	}
	toPushConfig(req.Body.Idx)
	return NewResponseSuccess(nil)
}

func updateProxyState(req *Request[ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.UpdateProxyState(req.Body.toDb())
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "update proxy configs failed")
	}
	toPushConfig(req.Body.Idx)
	return NewResponseSuccess(nil)
}

func getWebConfigs(req *Request[WebConfigInfo]) *Response {
	if req.Body.RefProxyId <= 0 {
		return NewResponseFail(errs.CodeSysErr, "refProxyId is empty")
	}
	wf, _ := getWebConfig(req.Body.RefProxyId)
	return NewResponseSuccess(wf)
}

func getWebConfig(refProxyId int) (*WebConfigInfo, bool) {
	item := sql.GetWebProxyConfig(refProxyId)
	if item == nil {
		return nil, false
	}
	wf := &WebConfigInfo{
		RefProxyId: item.RefProxyId,
		Id:         item.Id,
		CertId:     convertInt32ToPointer(item.CertId),
	}
	_ = json.Unmarshal([]byte(item.Proxy), &wf.Proxy)
	return wf, true
}

func addWebConfigs(req *Request[WebConfigInfo]) *Response {
	body := req.Body
	if body.RefProxyId <= 0 {
		return NewResponseFail(errs.CodeSysErr, "ProxyId is empty")
	}
	if body.Proxy == nil || len(body.Proxy) == 0 {
		return NewResponseFail(errs.CodeSysErr, "Http is empty")
	}
	config := sql.GetWebProxyConfig(body.RefProxyId)
	var err error
	if config == nil {
		err = sql.AddWebProxyConfig(body.toDb())
	} else {
		err = sql.UpdateWebProxyConfig(body.toDb())
	}
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "Add web configs failed")
	}
	//更新状态.
	oldConfig := sql.GetProxyConfigByIdNotState(body.RefProxyId)
	if oldConfig == nil {
		return NewResponseFail(errs.CodeSysErr, "Get proxy config failed")
	}
	if oldConfig.State != 1 {
		oldConfig.State = 1
		_ = sql.UpdateProxyState(oldConfig)
	}
	toPushConfig(body.RefProxyId)
	return NewResponseSuccess(nil)
}

func addProxyConfigs(req *Request[ProxyConfig]) *Response {
	body := req.Body
	if body.Name == "" {
		return NewResponseFail(errs.CodeSysErr, "name is empty")
	}
	if body.RemotePort < 10000 || body.RemotePort > 65535 {
		return NewResponseFail(errs.CodeSysErr, "port is invalid, the remote port range[10000-65535]")
	}
	if body.Protocol == "" {
		return NewResponseFail(errs.CodeSysErr, "protocol is empty")
	}
	if body.ProxyID == "" {
		return NewResponseFail(errs.CodeSysErr, "proxyId is empty")
	}
	if body.IsHttpOrHttps() {
		body.State = 0
	} else {
		body.State = 1
	}
	err := sql.AddProxyConfig(body.toDb())
	if err != nil {
		log.Error(err.Error())
		return NewResponseFail(errs.CodeSysErr, "add configs failed")
	}
	base.TunnelCfm.Push(body.ProxyID)
	return NewResponseSuccess(nil)
}

func toPushConfig(id int) {
	info := sql.GetProxyConfigByIdNotState(id)
	if info != nil {
		base.TunnelCfm.Push(info.ProxyID)
	}
}
