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

import "testing"

func TestFindRouteConflict(t *testing.T) {
	tests := []struct {
		name        string
		leftDomain  string
		leftPaths   []string
		rightDomain string
		rightPaths  []string
		want        bool
		wantDomain  string
		wantLeft    string
		wantRight   string
	}{
		{
			name:        "wildcard path overlaps specific path",
			leftDomain:  "*",
			leftPaths:   []string{"/*"},
			rightDomain: "*",
			rightPaths:  []string{"/webDev"},
			want:        true,
			wantDomain:  "*",
			wantLeft:    "/*",
			wantRight:   "/webDev",
		},
		{
			name:        "same domain different static path no overlap",
			leftDomain:  "example.com",
			leftPaths:   []string{"/webDev"},
			rightDomain: "example.com",
			rightPaths:  []string{"/webDev2"},
			want:        false,
		},
		{
			name:        "param path overlaps static path",
			leftDomain:  "example.com",
			leftPaths:   []string{"/user/:id"},
			rightDomain: "example.com",
			rightPaths:  []string{"/user/123"},
			want:        true,
			wantDomain:  "example.com",
			wantLeft:    "/user/:id",
			wantRight:   "/user/123",
		},
		{
			name:        "same path different explicit domains no overlap",
			leftDomain:  "a.example.com",
			leftPaths:   []string{"/api/user"},
			rightDomain: "b.example.com",
			rightPaths:  []string{"/api/user"},
			want:        false,
		},
		{
			name:        "wildcard domain overlaps exact domain",
			leftDomain:  "*.example.com",
			leftPaths:   []string{"/api/*"},
			rightDomain: "dev.example.com",
			rightPaths:  []string{"/api/user"},
			want:        true,
			wantDomain:  "dev.example.com",
			wantLeft:    "/api/*",
			wantRight:   "/api/user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindRouteConflict(tt.leftDomain, tt.leftPaths, tt.rightDomain, tt.rightPaths)
			if ok != tt.want {
				t.Fatalf("want conflict=%v, got %v", tt.want, ok)
			}
			if !tt.want {
				return
			}
			if got.Domain != tt.wantDomain || got.LeftPath != tt.wantLeft || got.RightPath != tt.wantRight {
				t.Fatalf("unexpected conflict: %+v", got)
			}
		})
	}
}
