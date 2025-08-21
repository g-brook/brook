package http

import (
	"net/http"

	"github.com/brook/common/utils"
)

var Routes []*RouteInfo

// ProxyConnectionFunction is a function that returns a net.Conn
type ProxyConnectionFunction func(proxyId string, reqId int64) (workConn *ProxyConnection, err error)

// RouteFunction is a function that returns a RouteInfo
type RouteFunction func(request *http.Request) (*RouteInfo, error)

// RouteInfo is a struct that holds information about a route
type RouteInfo struct {
	proxyId string

	matcher *utils.PathMatcher

	domain string

	getProxyConnection ProxyConnectionFunction
}

// AddRouteInfo adds a route to the Routes slice
func AddRouteInfo(proxyId string, domain string, paths []string, fun ProxyConnectionFunction) {
	info := &RouteInfo{
		proxyId:            proxyId,
		matcher:            utils.NewPathMatcher(),
		getProxyConnection: fun,
		domain:             domain,
	}
	for _, path := range paths {
		info.matcher.AddPathMatcher(path, info)
	}
	Routes = append(Routes, info)
}

// GetRouteInfo returns the RouteInfo for a given path
func GetRouteInfo(domain string, path string) *RouteInfo {
	var infos []*RouteInfo
	for _, info := range Routes {
		if !utils.MatchDomain(info.domain, domain) {
			continue
		}
		if info.matcher.Match(path).Matched {
			infos = append(infos, info)
		}
	}
	if infos != nil {
		return infos[0]
	}
	return nil
}
