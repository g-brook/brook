package api

import (
	"strconv"

	"github.com/brook/server/metrics"
	"github.com/brook/server/tunnel"
	"github.com/brook/server/web/errs"
)

func init() {
	RegisterRoute(NewRoute("/getServerInfo", "POST"), getServerInfo)
	RegisterRoute(NewRoute("/stopServer", "POST"), stopServer)
}

// GetServerInfo retrieves information about the server
// This function is designed to gather and return various details
// about the server's current status and configuration
func getServerInfo(req *Request[QueryServerInfo]) *Response {
	servers := metrics.M.GetServers()
	var v []*ServerInfo
	for _, item := range servers {
		v = append(v, &ServerInfo{
			Name:        item.Name(),
			Port:        strconv.Itoa(item.Port()),
			TunnelType:  item.Type(),
			TAG:         "",
			Connections: item.Connections(),
			Users:       item.Users(),
		})
	}

	return NewResponseSuccess(v)
}

func stopServer(req *Request[QueryServerInfo]) *Response {
	if req.Body.ProxyId == "" {
		return NewResponseFail(errs.CodeSysErr, "proxyId is empty")
	}
	for _, t := range metrics.M.GetServers() {
		b := t.Id() == req.Body.ProxyId
		if b {
			switch t := t.(type) {
			case tunnel.TunnelServer:
				t.(tunnel.TunnelServer).Shutdown()
			}
		}
	}
	return NewResponseSuccess(nil)

}
