package http

import "github.com/brook/common/utils"

var Routes []*RouteInfo

type RouteInfo struct {
	proxyId string

	matcher *utils.PathMatcher
}

func AddRouteInfo(proxyId string, paths []string) {
	info := &RouteInfo{
		proxyId: proxyId,
		matcher: utils.NewPathMatcher(),
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
