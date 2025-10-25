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
}

// getProxyConfigs retrieves configuration information from the database
// It takes a pointer to a Request with any type as parameter
// and returns a Response containing the configuration data or error information
func getProxyConfigs(*Request[any]) *Response {
	config := sql.QueryProxyConfig()
	if config == nil {
		return NewResponseSuccess(nil)
	}
	configMap := make(map[string]*sql.ProxyConfig)
	for _, cf := range config {
		configMap[cf.ProxyID] = cf
	}
	servers := metrics.M.GetServers()
	for _, server := range servers {
		proxyConfig, ok := configMap[server.Id()]
		if ok {
			proxyConfig.IsRunning = true
			proxyConfig.Runtime = server.Runtime().Format("2006-01-02 15:04:05")
			proxyConfig.Clients = server.Users()
		}
	}
	for _, v := range configMap {
		if (v.Protocol == "HTTP") || (v.Protocol == "HTTPS") {
			proxyConfig := sql.GetWebProxyConfig(v.Idx)
			v.IsExistWeb = proxyConfig != nil
		}
	}
	return NewResponseSuccess(config)
}

func delProxyConfig(req *Request[sql.ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.DelProxyConfig(req.Body.Idx)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete proxy configs failed")
	}
	return NewResponseSuccess(nil)
}

func updateProxyConfig(req *Request[sql.ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.UpdateProxyConfig(req.Body)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "update proxy configs failed")
	}
	return NewResponseSuccess(nil)
}

func updateProxyState(req *Request[sql.ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.UpdateProxyState(req.Body)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "update proxy configs failed")
	}
	return NewResponseSuccess(nil)
}

func getWebConfigs(req *Request[WebConfigInfo]) *Response {
	if req.Body.RefProxyId <= 0 {
		return NewResponseFail(errs.CodeSysErr, "refProxyId is empty")
	}
	item := sql.GetWebProxyConfig(req.Body.RefProxyId)
	if item == nil {
		return NewResponseSuccess(nil)
	}
	wf := &WebConfigInfo{
		RefProxyId: item.RefProxyId,
		Id:         item.Id,
		KeyFile:    item.KeyFile,
		CertFile:   item.CertFile,
	}
	_ = json.Unmarshal([]byte(item.Proxy), &wf.Proxy)
	return NewResponseSuccess(wf)
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
	info := sql.GetProxyConfigById(body.RefProxyId)
	if info != nil {
		base.CFM.Push(info.ProxyID)
	}
	return NewResponseSuccess(nil)
}

func addProxyConfigs(req *Request[sql.ProxyConfig]) *Response {
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
	if (body.Protocol == "HTTP") || (body.Protocol == "HTTPS") {
		body.State = 0
	} else {
		body.State = 1
	}
	err := sql.AddProxyConfig(body)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "add configs failed")
	}
	return NewResponseSuccess(nil)
}
