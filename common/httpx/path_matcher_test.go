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

package httpx_test

import (
	"testing"

	"github.com/g-brook/brook/common/httpx"
)

func TestNewPathMatcher(t *testing.T) {
	matcher := httpx.NewPathMatcher()
	if matcher == nil {
		t.Errorf("Expected non-nil PathMatcher")
		return
	}
	if matcher.Root == nil {
		t.Errorf("Expected non-nil root node")
	}
}

func TestAddAndGetPathMatcher(t *testing.T) {
	matcher := httpx.NewPathMatcher()
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
