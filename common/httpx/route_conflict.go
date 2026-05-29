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

package httpx

import "strings"

// RouteConflict describes one overlapping domain/path route pair.
type RouteConflict struct {
	Domain    string
	LeftPath  string
	RightPath string
}

// FindRouteConflict reports the first overlapping route between two domain/path rule sets.
func FindRouteConflict(leftDomain string, leftPaths []string, rightDomain string, rightPaths []string) (RouteConflict, bool) {
	conflictDomain, ok := overlapDomainValue(leftDomain, rightDomain)
	if !ok {
		return RouteConflict{}, false
	}
	for _, leftPath := range leftPaths {
		for _, rightPath := range rightPaths {
			if PathPatternOverlap(leftPath, rightPath) {
				return RouteConflict{
					Domain:    conflictDomain,
					LeftPath:  leftPath,
					RightPath: rightPath,
				}, true
			}
		}
	}
	return RouteConflict{}, false
}

// DomainPatternOverlap reports whether two domain patterns can match the same host.
func DomainPatternOverlap(left string, right string) bool {
	_, ok := overlapDomainValue(left, right)
	return ok
}

// PathPatternOverlap reports whether two path patterns can match the same request path.
func PathPatternOverlap(left string, right string) bool {
	leftParts := splitRoutePath(left)
	rightParts := splitRoutePath(right)
	return matchRoutePatternOverlap(leftParts, rightParts, 0, 0)
}

func overlapDomainValue(left string, right string) (string, bool) {
	left = normalizeDomainPattern(left)
	right = normalizeDomainPattern(right)
	if left != right && isExplicitDomainPattern(left) && isExplicitDomainPattern(right) {
		return "", false
	}
	if MatchDomain(left, right) {
		return DisplayRouteValue(right, left), true
	}
	if MatchDomain(right, left) {
		return DisplayRouteValue(left, right), true
	}
	return "", false
}

func normalizeDomainPattern(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}

func isExplicitDomainPattern(domain string) bool {
	if domain == "" || domain == "*" {
		return false
	}
	return !strings.HasPrefix(domain, "*.")
}

// DisplayRouteValue returns the preferred printable route value for messages.
func DisplayRouteValue(value string, fallback string) string {
	if value != "" {
		return value
	}
	if fallback != "" {
		return fallback
	}
	return "*"
}

func matchRoutePatternOverlap(left []string, right []string, leftIndex int, rightIndex int) bool {
	if leftIndex == len(left) && rightIndex == len(right) {
		return true
	}
	if leftIndex == len(left) {
		return rightIndex < len(right) && isWildcardSegment(right[rightIndex])
	}
	if rightIndex == len(right) {
		return leftIndex < len(left) && isWildcardSegment(left[leftIndex])
	}
	if isWildcardSegment(left[leftIndex]) || isWildcardSegment(right[rightIndex]) {
		return true
	}
	if !routeSegmentOverlap(left[leftIndex], right[rightIndex]) {
		return false
	}
	return matchRoutePatternOverlap(left, right, leftIndex+1, rightIndex+1)
}

func routeSegmentOverlap(left string, right string) bool {
	if left == right {
		return true
	}
	if isParamSegment(left) || isParamSegment(right) {
		return true
	}
	return false
}

func isParamSegment(segment string) bool {
	return strings.HasPrefix(segment, ":")
}

func isWildcardSegment(segment string) bool {
	return strings.HasPrefix(segment, "*")
}

func splitRoutePath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
