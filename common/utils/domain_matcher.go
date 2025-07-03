package utils

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
