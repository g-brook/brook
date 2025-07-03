package utils

import (
	"testing"
)

func TestMatchDomain(t *testing.T) {
	testCases := []struct {
		name        string
		routeDomain string
		reqDomain   string
		expected    bool
	}{
		{
			name:        "Empty route domain",
			routeDomain: "",
			reqDomain:   "example.com",
			expected:    true,
		},
		{
			name:        "Exact match",
			routeDomain: "example.com",
			reqDomain:   "example.com",
			expected:    true,
		},
		{
			name:        "No match",
			routeDomain: "example.com",
			reqDomain:   "test.example.com",
			expected:    false,
		},
		{
			name:        "Wildcard match 1 level",
			routeDomain: "*.example.com",
			reqDomain:   "test.example.com",
			expected:    true,
		},
		{
			name:        "Wildcard match multiple levels",
			routeDomain: "*.example.com",
			reqDomain:   "test.sub.example.com",
			expected:    true,
		},
		{
			name:        "Wildcard no match",
			routeDomain: "*.example.com",
			reqDomain:   "example.com",
			expected:    false,
		},
		{
			name:        "Wildcard different domain",
			routeDomain: "*.example.com",
			reqDomain:   "test.another.com",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MatchDomain(tc.routeDomain, tc.reqDomain)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v", tc.expected, result)
			}
		})
	}
}
