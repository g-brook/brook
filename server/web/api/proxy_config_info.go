package api

import (
	"github.com/brook/server/web/errs"
	"github.com/brook/server/web/sql"
)

func init() {
	RegisterRoute(NewRoute("/getProxyConfigs", "POST"), getProxyConfigs)
	RegisterRoute(NewRoute("/addProxyConfigs", "POST"), addProxyConfigs)
	RegisterRoute(NewRoute("/delProxyConfigs", "POST"), delProxyConfig)
}

// getProxyConfigs retrieves configuration information from the database
// It takes a pointer to a Request with any type as parameter
// and returns a Response containing the configuration data or error information
func getProxyConfigs(*Request[any]) *Response {
	config := sql.QueryProxyConfig()
	if config == nil {
		return NewResponseSuccess(nil)
	}
	return NewResponseSuccess(config)
}

func delProxyConfig(req *Request[sql.ProxyConfig]) *Response {
	if req.Body.Idx <= 0 {
		return NewResponseFail(errs.CodeSysErr, "idx is empty")
	}
	err := sql.DelProxyConfig(req.Body.Idx)
	if err != nil {
		return NewResponseFail(errs.CodeSysErr, "delete proxy config failed")
	}
	return NewResponseSuccess(nil)
}

func addProxyConfigs(req *Request[sql.ProxyConfig]) *Response {
	body := req.Body
	if body.Name == "" {
		return NewResponseFail(errs.CodeSysErr, "name is empty")
	}
	if body.RemotePort < 30000 || body.RemotePort > 65535 {
		return NewResponseFail(errs.CodeSysErr, "port is invalid, the remote port range[30000-65535]")
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
		return NewResponseFail(errs.CodeSysErr, "add config failed")
	}
	return NewResponseSuccess(nil)
}
