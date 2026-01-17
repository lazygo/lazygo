package server

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Common struct for Server & Group.
type common struct {
	add func(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc)
}

// Connect registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Connect(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodConnect, path, h, m...)
}

// Delete registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (cm *common) Delete(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodDelete, path, h, m...)
}

// Get registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (cm *common) Get(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodGet, path, h, m...)
}

// Head registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Head(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodHead, path, h, m...)
}

// Options registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Options(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodOptions, path, h, m...)
}

// Patch registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Patch(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodPatch, path, h, m...)
}

// Post registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Post(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodPost, path, h, m...)
}

// Put registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Put(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodPut, path, h, m...)
}

// Trace registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) Trace(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(http.MethodTrace, path, h, m...)
}

// WebSocket registers a new WEBSOCKET route for a path with matching handler in the
// router with optional route-level middleware.
func (cm *common) WebSocket(path string, h HandlerFunc, m ...MiddlewareFunc) {
	cm.add(MethodWebSocket, path, h, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (cm *common) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		cm.add(m, path, handler, middleware...)
	}
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (cm *common) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		cm.add(m, path, handler, middleware...)
	}
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (cm *common) Static(prefix, root string) {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}

	h := func(c Context) error {
		ps, _ := c.Param("*")
		p, err := url.PathUnescape(ps)
		if err != nil {
			return err
		}

		name := filepath.Join(root, filepath.Clean("/"+p)) // "/"+ for security
		fi, err := os.Stat(name)
		if err != nil {
			// The access path does not exist
			return NotFoundHandler(c)
		}

		// If the request is for a directory and does not end with "/"
		p = c.Request().URL.Path // path must not be empty.
		if fi.IsDir() && p[len(p)-1] != '/' {
			// Redirect to ends with "/"
			return c.Redirect(http.StatusMovedPermanently, p+"/")
		}
		return c.File(name)
	}
	if prefix == "/" {
		cm.add(http.MethodGet, prefix+"*", h)
	}
	cm.add(http.MethodGet, prefix+"/*", h)
}

// File registers a new route with path to serve a static file with optional route-level middleware.
func (cm *common) File(path, file string, m ...MiddlewareFunc) {
	cm.add(http.MethodGet, path, func(c Context) error {
		return c.File(file)
	}, m...)
}
