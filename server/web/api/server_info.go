package api

import (
	"strconv"

	"github.com/brook/server/metrics"
)

func init() {
	RegisterRoute(NewRoute("/getServerInfo", "POST"), getServerInfo)
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
