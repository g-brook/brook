package http

import (
	"github.com/brook/common/utils"
	"net"
	"net/http"
)

var Routes []*RouteInfo

type ProxyConnectionFunction func(proxyId string) (net.Conn, error)

type RouteFunction func(request *http.Request) (*RouteInfo, error)

type RouteInfo struct {
	proxyId string

	matcher *utils.PathMatcher

	getProxyConnection ProxyConnectionFunction
}

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

func GetRouteInfo(path string) *RouteInfo {
	for _, info := range Routes {
		if info.matcher.Match(path).Matched {
			return info
		}
	}
	return nil
}
