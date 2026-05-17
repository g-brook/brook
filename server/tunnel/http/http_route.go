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
	"strings"
	"sync"

	"github.com/g-brook/brook/common/httpx"
	"github.com/g-brook/brook/common/log"
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

	paths []string

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
		paths:              append([]string(nil), paths...),
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
	var selected *RouteInfo
	bestScore := -1
	for _, info := range routes {
		if !httpx.MatchDomain(info.domain, domain) {
			continue
		}
		if info.matcher.Match(path).Matched {
			matchScore := routeScore(info, domain, path)
			if matchScore > bestScore {
				bestScore = matchScore
				selected = info
			}
		}
	}
	if selected == nil {
		log.Warn("No route info for domain %s", domain)
	}
	return selected
}

func routeScore(info *RouteInfo, domain string, path string) int {
	score := 0
	if info.domain == domain {
		score += 10000
	} else {
		score += len(info.domain) * 100
	}

	bestPathScore := 0
	for _, pattern := range info.paths {
		if !patternMatch(pattern, path) {
			continue
		}
		pathScore := patternSpecificity(pattern)
		if pathScore > bestPathScore {
			bestPathScore = pathScore
		}
	}
	return score + bestPathScore
}

func patternMatch(pattern string, path string) bool {
	patternParts := splitRoutePattern(pattern)
	pathParts := splitRoutePattern(path)
	return matchParts(patternParts, pathParts)
}

func matchParts(patternParts []string, pathParts []string) bool {
	if len(patternParts) == 0 {
		return len(pathParts) == 0
	}
	for i := 0; i < len(patternParts); i++ {
		if i >= len(pathParts) {
			return false
		}
		segment := patternParts[i]
		switch {
		case segment == "":
			continue
		case segment[0] == '*':
			return true
		case segment[0] == ':':
			continue
		default:
			if segment != pathParts[i] {
				return false
			}
		}
	}
	return len(pathParts) == len(patternParts)
}

func patternSpecificity(pattern string) int {
	segments := splitRoutePattern(pattern)
	score := 0
	for _, segment := range segments {
		switch {
		case segment == "":
		case segment[0] == '*':
			score += 1
		case segment[0] == ':':
			score += 10
		default:
			score += 100 + len(segment)
		}
	}
	return score + len(segments)
}

func splitRoutePattern(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
