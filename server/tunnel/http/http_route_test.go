package http

import "testing"

func TestPatternMatch(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"/user/profile", "/user/profile", true},
		{"/user/:id", "/user/123", true},
		{"/user/:id", "/user/123/extra", false},
		{"/user/*rest", "/user/123/extra", true},
		{"/", "/", true},
		{"/", "/x", false},
	}

	for _, tc := range tests {
		if got := patternMatch(tc.pattern, tc.path); got != tc.want {
			t.Fatalf("patternMatch(%q, %q) = %v, want %v", tc.pattern, tc.path, got, tc.want)
		}
	}
}

func TestGetRouteInfoPrefersMoreSpecificPattern(t *testing.T) {
	RouteClean()
	AddRouteInfo("wild", "example.com", []string{"/user/*rest"}, nil)
	AddRouteInfo("param", "example.com", []string{"/user/:id"}, nil)
	AddRouteInfo("static", "example.com", []string{"/user/profile"}, nil)

	info := GetRouteInfo("example.com", "/user/profile")
	if info == nil {
		t.Fatal("expected route info")
	}
	if info.httpId != "static" {
		t.Fatalf("expected static route, got %q", info.httpId)
	}
}

func TestGetRouteInfoPrefersExactDomain(t *testing.T) {
	RouteClean()
	AddRouteInfo("wildcard-domain", "*.example.com", []string{"/api/status"}, nil)
	AddRouteInfo("exact-domain", "api.example.com", []string{"/api/status"}, nil)

	info := GetRouteInfo("api.example.com", "/api/status")
	if info == nil {
		t.Fatal("expected route info")
	}
	if info.httpId != "exact-domain" {
		t.Fatalf("expected exact-domain route, got %q", info.httpId)
	}
}
