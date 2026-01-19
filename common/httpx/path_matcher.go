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
	"strings"

	"github.com/brook/common/log"
)

type PathMatcher struct {
	Root *node
}

type node struct {
	segment    string
	children   []*node
	isParam    bool
	isWildcard bool
	handler    interface{}
}

func NewPathMatcher(handler interface{}) *PathMatcher {
	return &PathMatcher{
		Root: &node{
			segment:    "",
			children:   make([]*node, 0),
			isParam:    false,
			isWildcard: false,
			handler:    handler,
		},
	}
}

// AddPathMatcher adds a new path matcher to the tree
// It takes a path string and a handler interface as parameters
func (r *PathMatcher) AddPathMatcher(path string, handler interface{}) {
	if r.Root == nil {
		log.Info("AddPathMatcher: root node is nil")
		return
	}
	// Split the path into parts
	parts := splitPath(path)
	// Start from the root node
	cur := r.Root
	// Traverse through each part of the path
	for _, part := range parts {
		var child *node
		// Check if current part already exists as a child node
		for _, c := range cur.children {
			if c.segment == part {
				child = c
				break
			}
		}
		// If child doesn't exist, create a new node
		if child == nil {
			child = &node{
				segment:    part,                         // The path segment
				children:   make([]*node, 0),             // Initialize empty children slice
				isParam:    strings.HasPrefix(part, ":"), // Check if it's a parameter
				isWildcard: strings.HasPrefix(part, "*"), // Check if it's a wildcard
				handler:    handler,
			}
			// Add the new child to current node's children
			cur.children = append(cur.children, child)
		}
		// Move to the child node for next iteration
		cur = child
	}
	// Set the handler for the final node
	cur.handler = handler
}

// Match checks if a given path matches the routing tree
// It returns a MatchResult indicating whether the path was matched and any extracted parameters
func (r *PathMatcher) Match(path string) MatchResult {
	// Split the input path into its components
	parts := splitPath(path)
	// Initialize a map to store any parameters extracted from the path
	params := make(map[string]string)
	// Attempt to find a matching node in the routing tree
	node, ok := matchNode(r.Root, parts, params)
	// If a match was found and the node has a handler, return success
	if ok && node.handler != nil {
		return MatchResult{
			Matched: true,
		}
	}
	// Otherwise return failure
	return MatchResult{
		Matched: false,
	}
}

// matchNode is a recursive function that matches a URL path against a tree of nodes.
// It checks if the given path parts match the route structure and extracts any parameters.
//
// Parameters:
//
//	n - The current node in the route tree to match against
//	parts - The remaining parts of the URL path to match
//	params - A map to store any extracted parameters from the URL
//
// Returns:
//
//	*node - The matched node if successful, nil otherwise
//	bool - True if a match was found, false otherwise
func matchNode(n *node, parts []string, params map[string]string) (*node, bool) {
	// Base case: if there are no more parts to match, we've successfully matched the path
	if len(parts) == 0 {
		return n, true
	}

	// Get the first part of the path to match against current node's children
	part := parts[0]
	// Iterate through all children of the current node
	for _, child := range n.children {
		switch {
		// Case 1: Exact match with current segment
		case child.segment == part:
			// Recursively match the remaining parts
			if res, ok := matchNode(child, parts[1:], params); ok {
				return res, true
			}
		// Case 2: Parameter match (segment starts with ':')
		case child.isParam:
			// Extract parameter name (without the leading ':')
			key := child.segment[1:]
			// Store the parameter value in the params map
			params[key] = part
			// Recursively match the remaining parts
			if res, ok := matchNode(child, parts[1:], params); ok {
				return res, true
			}
			// If matching fails, remove the parameter from the map
			delete(params, key)
		// Case 3: Wildcard match (segment starts with '*')
		case child.isWildcard:
			// Extract wildcard parameter name (without the leading '*')
			key := child.segment[1:]
			// Join all remaining parts to form the wildcard parameter value
			params[key] = strings.Join(parts, "/")
			// Return the wildcard node immediately
			return child, true
		}
	}
	// If no match is found among all children, return nil and false
	return nil, false
}

type MatchResult struct {
	Matched bool
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	parts := strings.Split(path, "/")

	// 处理根路径的特殊情况
	if len(parts) == 1 && parts[0] == "" {
		return []string{}
	}

	return parts
}
