package utils_test

import (
	"github.com/brook/common/utils"
	"testing"
)

func TestNewPathMatcher(t *testing) {
	matcher := utils.NewPathMatcher()
	if matcher == nil {
		t.Errorf("Expected non-nil PathMatcher")
	}
	if matcher.root == nil {
		t.Errorf("Expected non-nil root node")
	}
}

func TestAddAndGetPathMatcher(t *testing) {
	matcher := utils.NewPathMatcher()
	testCases := []struct {
		path    string
		handler interface{}
	}{
		{"/", "index"},
		{"/user", "user"},
		{"/user/:id", "user"},
		{"/user/:id/:name", "user"},
		{"/user/:id/:name/:age", "user"},
		{"/user/:id/:name/:age/:sex", "user"},
		{"/user/:id/:name/:age/:sex/:address", "user"},
		{"/index/*", false},
	}

	// Test adding paths
	for _, tc := range testCases {
		matcher.AddPathMatcher(tc.path, tc.handler)
	}

	// Test matching paths
	matchResults := []struct {
		input    string
		expected bool
		testName string
	}{
		{"/index/aaa/bbb", true, "Wildcard match"},
		{"/ok/aaa/bbb", false, "No match"},
		{"/user/aaa/bbb", true, "Parametric match"},
	}

	for _, result := range matchResults {
		t.Run(result.testName, func(t *testing.T) {
			match := matcher.Match(result.input)
			if match.Matched != result.expected {
				t.Errorf("For input %s, expected %v but got %v", result.input, result.expected, match.Matched)
			}
		})
	}
}
