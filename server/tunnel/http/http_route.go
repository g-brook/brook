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

package http

import (
	"net"
	"net/http"
	"sync"

	"github.com/brook/common/httpx"
)

var routes []*RouteInfo

var lock sync.RWMutex

// ProxyConnectionFunction is a function that returns a net.Conn
type ProxyConnectionFunction func(httpId string) (workConn net.Conn, err error)

// RouteFunction is a function that returns a RouteInfo
type RouteFunction func(request *http.Request) (*RouteInfo, error)

// RouteInfo is a struct that holds information about a route
type RouteInfo struct {
	httpId string

	matcher *httpx.PathMatcher

	domain string

	getProxyConnection ProxyConnectionFunction
}

// AddRouteInfo adds a route to the routes slice
func AddRouteInfo(httpId string, domain string, paths []string, fun ProxyConnectionFunction) {
	lock.Lock()
	defer lock.Unlock()
	info := &RouteInfo{
		httpId:             httpId,
		getProxyConnection: fun,
		domain:             domain,
	}
	info.matcher = httpx.NewPathMatcher(info)
	for _, path := range paths {
		info.matcher.AddPathMatcher(path, info)
	}
	routes = append(routes, info)
}

func RouteClean() {
	lock.Lock()
	defer lock.Unlock()
	routes = routes[:0]
}

// GetRouteInfo returns the RouteInfo for a given path
func GetRouteInfo(domain string, path string) *RouteInfo {
	lock.RLock()
	defer lock.RUnlock()
	var infos []*RouteInfo
	for _, info := range routes {
		if !httpx.MatchDomain(info.domain, domain) {
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
