package utils

import "strings"

type PathMatcher struct {
	root *node
}

type node struct {
	segment    string
	children   []*node
	isParam    bool
	isWildcard bool
	handler    interface{}
}

func NewPathMatcher() *PathMatcher {
	return &PathMatcher{
		root: &node{},
	}
}

func (r *PathMatcher) AddPathMatcher(path string, handler interface{}) {
	parts := splitPath(path)
	cur := r.root
	for _, part := range parts {
		var child *node
		for _, c := range cur.children {
			if c.segment == part {
				child = c
				break
			}
		}
		if child == nil {
			child = &node{
				segment:    part,
				children:   make([]*node, 0),
				isParam:    strings.HasPrefix(part, ":"),
				isWildcard: strings.HasPrefix(part, "*"),
			}
			cur.children = append(cur.children, child)
		}
		cur = child
	}
	cur.handler = handler
}

func (r *PathMatcher) Match(path string) MatchResult {
	parts := splitPath(path)
	params := make(map[string]string)
	node, ok := matchNode(r.root, parts, params)
	if ok && node.handler != nil {
		return MatchResult{
			Matched: true,
		}
	}
	return MatchResult{
		Matched: false,
	}
}

func matchNode(n *node, parts []string, params map[string]string) (*node, bool) {
	if len(parts) == 0 {
		return n, true
	}

	part := parts[0]
	for _, child := range n.children {
		switch {
		case child.segment == part:
			if res, ok := matchNode(child, parts[1:], params); ok {
				return res, true
			}
		case child.isParam:
			key := child.segment[1:]
			params[key] = part
			if res, ok := matchNode(child, parts[1:], params); ok {
				return res, true
			}
			delete(params, key)
		case child.isWildcard:
			key := child.segment[1:]
			params[key] = strings.Join(parts, "/")
			return child, true
		}
	}
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
	return strings.Split(path, "/")
}
