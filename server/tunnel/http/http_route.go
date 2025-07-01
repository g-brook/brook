package http

import (
	"github.com/brook/common/utils"
	"net/http"
)

var Routes []*RouteInfo

// ProxyConnectionFunction is a function that returns a net.Conn
type ProxyConnectionFunction func(proxyId string, reqId string) (workConn *ProxyConnection, err error)

// RouteFunction is a function that returns a RouteInfo
type RouteFunction func(request *http.Request) (*RouteInfo, error)

// RouteInfo is a struct that holds information about a route
type RouteInfo struct {
	proxyId string

	matcher *utils.PathMatcher

	getProxyConnection ProxyConnectionFunction
}

// AddRouteInfo adds a route to the Routes slice
func AddRouteInfo(proxyId string, paths []string, fun ProxyConnectionFunction) {
	info := &RouteInfo{
		proxyId:            proxyId,
		matcher:            utils.NewPathMatcher(),
		getProxyConnection: fun,
	}
	for _, path := range paths {
		info.matcher.AddPathMatcher(path, info)
	}
	Routes = append(Routes, info)
}

// GetRouteInfo returns the RouteInfo for a given path
func GetRouteInfo(path string) *RouteInfo {
	for _, info := range Routes {
		if info.matcher.Match(path).Matched {
			return info
		}
	}
	return nil
}
