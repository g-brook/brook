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
