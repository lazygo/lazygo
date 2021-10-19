package server

import "strings"

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middleware or functionality that should be separate
	// from the parent engine instance while still inheriting from it.
	Group struct {
		common
		prefix     string
		middleware []MiddlewareFunc
		server     *Server
	}
)

func newGroup(prefix string, s *Server) *Group {
	g := &Group{prefix: prefix, server: s}
	g.common.add = g.Add
	return g
}

// Use implements `Server#Use()` for sub-routes within the Group.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
	// if len(g.middleware) == 0 {
	// 	return
	// }
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
	// g.Any("", NotFoundHandler)
	// g.Any("/*", NotFoundHandler)
}

// Add implements `Server#Add()` for sub-routes within the Group.
func (g *Group) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	// Combine into a new slice to avoid accidentally passing the same slice for
	// multiple routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	g.server.Add(method, g.concat(g.prefix, path), handler, m...)
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	m := make([]MiddlewareFunc, 0, len(g.middleware)+len(middleware))
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.server.Group(g.concat(g.prefix, prefix), m...)
}

func (g *Group) concat(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return strings.TrimRight(a, "/\\") + "/" + strings.TrimLeft(b, "/\\")
}
