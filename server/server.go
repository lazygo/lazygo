package server

import (
	stdContext "context"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
)

type (
	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// HTTPOKHandler is a centralized HTTP ok handler.
	HTTPOKHandler func(any, Context) error

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(Context) error

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]any
)

// Server is the top-level framework instance.
type Server struct {
	common
	premiddleware    []MiddlewareFunc
	middleware       []MiddlewareFunc
	maxParam         *int
	router           *Router
	notFoundHandler  HandlerFunc
	pool             sync.Pool
	Http             *http.Server
	Listener         net.Listener
	Debug            bool
	HTTPErrorHandler HTTPErrorHandler
	HTTPOKHandler    HTTPOKHandler
	Logger           *log.Logger
	ListenerNetwork  string
}

var (
	methods = [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		MethodPropfind,
		http.MethodPut,
		http.MethodTrace,
		MethodReport,
		MethodWebSocket,
	}
)

// New creates an instance of Server.
func New() (s *Server) {
	s = &Server{
		Http:            new(http.Server),
		maxParam:        new(int),
		ListenerNetwork: "tcp",
	}
	s.common.add = s.Add
	s.Http.Handler = s
	s.HTTPOKHandler = s.DefaultHTTPOKHandler
	s.HTTPErrorHandler = s.DefaultHTTPErrorHandler
	s.pool.New = func() any {
		return s.NewContext(nil, nil)
	}
	s.router = NewRouter(s)
	return
}

// NewContext returns a Context instance.
func (s *Server) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request:        r,
		responseWriter: NewResponseWriter(w),
		store:          make(Map),
		server:         s,
		pvalues:        make([]string, *s.maxParam),
		handler:        NotFoundHandler,
	}
}

// Router returns the default router.
func (s *Server) Router() *Router {
	return s.router
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (s *Server) DefaultHTTPErrorHandler(err error, c Context) {
	he, ok := err.(*HTTPError)
	if ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = &HTTPError{
			Code:    http.StatusInternalServerError,
			Errno:   http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	// Issue #1426
	code := he.Code
	message := he.Message
	if m, ok := he.Message.(string); ok {
		if s.Debug {
			message = Map{"code": code, "message": m, "error": err.Error()}
		} else {
			message = Map{"code": code, "message": m}
		}
	}

	// Send response
	if !c.ResponseWriter().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(he.Code)
		} else {
			err = c.JSON(code, message)
		}
		if err != nil {
			panic(err)
		}
	}
}

// DefaultHTTPOKHandler is the default HTTP ok handler. It sends a JSON response
// with status code.
func (s *Server) DefaultHTTPOKHandler(data any, c Context) error {

	message := Map{"errno": 0, "data": data}

	var err error
	// Send response
	if !c.ResponseWriter().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(http.StatusOK)
		} else {
			err = c.JSON(http.StatusOK, message)
		}
	}
	return err
}

// Pre adds middleware to the chain which is run before router.
func (s *Server) Pre(middleware ...MiddlewareFunc) {
	s.premiddleware = append(s.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (s *Server) Use(middleware ...MiddlewareFunc) {
	s.middleware = append(s.middleware, middleware...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (s *Server) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	s.router.Add(method, path, func(c Context) error {
		h := applyMiddleware(handler, middleware...)
		return h(c)
	})
}

// Group creates a new router group with prefix and optional group-level middleware.
func (s *Server) Group(prefix string, m ...MiddlewareFunc) *Group {
	g := newGroup(prefix, s)
	g.Use(m...)
	return g
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (s *Server) AcquireContext() Context {
	return s.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (s *Server) ReleaseContext(c Context) {
	s.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire context
	c := s.pool.Get().(*context)
	// Release context
	defer s.pool.Put(c)
	c.reset(r, w)

	ctx := Context(c)

	h := func(c Context) error {
		// premiddleware 在路由查找之前加载
		// 可以在premiddleware中处理url路径等参数以改变路由查找的行为
		s.router.Find(r.Method, r.URL.EscapedPath(), ctx.c())
		h := c.Handler()
		h = applyMiddleware(h, s.middleware...)
		ctx = c
		return h(c)
	}
	h = applyMiddleware(h, s.premiddleware...)

	defer func() {
		rec := recover()
		if rec != nil {
			fmt.Println("panic:", rec)
			debug.PrintStack()
		}
	}()
	// Execute chain
	if err := h(c); err != nil {
		s.HTTPErrorHandler(err, ctx)
	}
}

// Start starts an HTTP server.
func (s *Server) Start(ctx stdContext.Context, address string) error {
	s.Http.Addr = address
	return s.StartServer(ctx, s.Http)
}

// StartServer starts a custom http server.
func (s *Server) StartServer(ctx stdContext.Context, h *http.Server) (err error) {
	// Setup
	if s.Logger != nil {
		h.ErrorLog = s.Logger
	}
	h.Handler = s
	h.BaseContext = func(l net.Listener) stdContext.Context {
		return ctx
	}

	if s.Listener == nil {
		s.Listener, err = newListener(h.Addr, s.ListenerNetwork)
		if err != nil {
			return err
		}
	}
	return h.Serve(s.Listener)
}

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (s *Server) Close() error {
	return s.Http.Close()
}

// Shutdown stops the server gracefully.
// It internally calls `http.Server#Shutdown()`.
func (s *Server) Shutdown(ctx stdContext.Context) error {
	return s.Http.Shutdown(ctx)
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	if len(middleware) == 0 {
		return h
	}
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
