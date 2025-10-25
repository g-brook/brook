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

// MatchDomain match domain.
func MatchDomain(routeDomain, reqDomain string) bool {
	if routeDomain == "" {
		return true
	}

	// Preprocess domain names
	var routeLabels, reqLabels []string
	var hasPrefix bool
	if strings.HasPrefix(routeDomain, "*.") {
		hasPrefix = true
		routeLabels = strings.Split(routeDomain[2:], ".") // Skip "*. " prefix
	} else {
		routeLabels = strings.Split(routeDomain, ".")
	}
	if hasPrefix {
		reqLabels = strings.Split(reqDomain, ".")[1:]
	} else {
		reqLabels = strings.Split(reqDomain, ".")
	}

	if len(routeLabels) > 1 {
		// Check if there are enough subdomains
		if len(reqLabels) < len(routeLabels) {
			return false
		}
		// For wildcard domains, compare parent domains
		if hasPrefix {
			// Compare each level of domain name after the wildcard
			for i := 0; i < len(routeLabels); i++ {
				if reqLabels[len(reqLabels)-len(routeLabels)+i] != routeLabels[i] {
					return false
				}
			}
		} else {
			// Directly compare complete domain names
			return routeDomain == reqDomain
		}
	} else {
		// Simple case: directly compare domain names or check wildcard
		return routeDomain == reqDomain || (strings.HasPrefix(routeDomain, "*.") && reqLabels[len(reqLabels)-1] == routeLabels[0])
	}
	return true
}
