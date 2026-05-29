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

	"github.com/g-brook/brook/common/httpx"
)

// ProxyConnectionFunction is a function that returns a net.Conn.
type ProxyConnectionFunction func(httpId string) (workConn net.Conn, err error)

// RouteFunction is a function that returns a RouteInfo.
type RouteFunction func(request *http.Request) (*RouteInfo, error)

// RouteInfo holds route matching and backend connection metadata.
type RouteInfo struct {
	httpId             string
	matcher            *httpx.PathMatcher
	domain             string
	getProxyConnection ProxyConnectionFunction
}

// Router stores routes for a single HTTP tunnel server instance.
type Router struct {
	lock   sync.RWMutex
	routes []*RouteInfo
}

// NewRouter creates an isolated router for one server instance.
func NewRouter() *Router {
	return &Router{routes: make([]*RouteInfo, 0)}
}

// AddRouteInfo adds a route to this router.
func (r *Router) AddRouteInfo(httpId string, domain string, paths []string, fun ProxyConnectionFunction) {
	if r == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	info := &RouteInfo{httpId: httpId, getProxyConnection: fun, domain: domain}
	info.matcher = httpx.NewPathMatcher(info)
	for _, path := range paths {
		info.matcher.AddPathMatcher(path, info)
	}
	r.routes = append(r.routes, info)
}

// Clean removes all routes from this router.
func (r *Router) Clean() {
	if r == nil {
		return
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.routes = r.routes[:0]
}

// GetRouteInfo returns the first matching route from this router.
func (r *Router) GetRouteInfo(domain string, path string) *RouteInfo {
	if r == nil {
		return nil
	}
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, info := range r.routes {
		if !httpx.MatchDomain(info.domain, domain) {
			continue
		}
		if info.matcher.Match(path).Matched {
			return info
		}
	}
	return nil
}
