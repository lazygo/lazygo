package engine

import (
	stdContext "context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

type (
	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to serve HTTP requests.
	HandlerFunc func(Context) error

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}

	// 控制器
	Controller struct{}
)

// Engine is the top-level framework instance.
type Engine struct {
	common
	premiddleware    []MiddlewareFunc
	middleware       []MiddlewareFunc
	maxParam         *int
	router           *Router
	notFoundHandler  HandlerFunc
	pool             sync.Pool
	Server           *http.Server
	Listener         net.Listener
	Debug            bool
	HTTPErrorHandler HTTPErrorHandler
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
		PROPFIND,
		http.MethodPut,
		http.MethodTrace,
		REPORT,
	}
)

// New creates an instance of Engine.
func New() (e *Engine) {
	e = &Engine{
		Server:          new(http.Server),
		maxParam:        new(int),
		ListenerNetwork: "tcp",
	}
	e.common.add = e.Add
	e.Server.Handler = e
	e.HTTPErrorHandler = e.DefaultHTTPErrorHandler
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}
	e.router = NewRouter(e)
	return
}

// NewContext returns a Context instance.
func (e *Engine) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request:        r,
		responseWriter: NewResponseWriter(w),
		store:          make(Map),
		engine:         e,
		pvalues:        make([]string, *e.maxParam),
		handler:        NotFoundHandler,
	}
}

// Router returns the default router.
func (e *Engine) Router() *Router {
	return e.router
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (e *Engine) DefaultHTTPErrorHandler(err error, c Context) {
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
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	// Issue #1426
	code := he.Code
	message := he.Message
	if m, ok := he.Message.(string); ok {
		if e.Debug {
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

// Pre adds middleware to the chain which is run before router.
func (e *Engine) Pre(middleware ...MiddlewareFunc) {
	e.premiddleware = append(e.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (e *Engine) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Engine) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	e.router.Add(method, path, func(c Context) error {
		h := applyMiddleware(handler, middleware...)
		return h(c)
	})
}

// Group creates a new router group with prefix and optional group-level middleware.
func (e *Engine) Group(prefix string, m ...MiddlewareFunc) *Group {
	g := newGroup(prefix, e)
	g.Use(m...)
	return g
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (e *Engine) AcquireContext() Context {
	return e.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (e *Engine) ReleaseContext(c Context) {
	e.pool.Put(c)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire context
	c := e.pool.Get().(*context)
	c.reset(r, w)

	h := NotFoundHandler

	if e.premiddleware == nil {
		e.router.Find(r.Method, r.URL.EscapedPath(), c)
		h = c.Handler()
		h = applyMiddleware(h, e.middleware...)
	} else {
		h = func(c Context) error {
			e.router.Find(r.Method, r.URL.EscapedPath(), c)
			h := c.Handler()
			h = applyMiddleware(h, e.middleware...)
			return h(c)
		}
		h = applyMiddleware(h, e.premiddleware...)
	}

	defer func() {
		rec := recover()
		if rec != nil {
			fmt.Println(rec)
			e.HTTPErrorHandler(fmt.Errorf("%v", rec), c)
		}
	}()
	// Execute chain
	if err := h(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}

	// Release context
	e.pool.Put(c)
}

// Start starts an HTTP server.
func (e *Engine) Start(address string) error {
	e.Server.Addr = address
	return e.StartServer(e.Server)
}

// StartServer starts a custom http server.
func (e *Engine) StartServer(s *http.Server) (err error) {
	// Setup
	if e.Logger != nil {
		s.ErrorLog = e.Logger
	}
	s.Handler = e

	if e.Listener == nil {
		e.Listener, err = newListener(s.Addr, e.ListenerNetwork)
		if err != nil {
			return err
		}
	}
	return s.Serve(e.Listener)
}

// Close immediately stops the server.
// It internally calls `http.Server#Close()`.
func (e *Engine) Close() error {
	return e.Server.Close()
}

// Shutdown stops the server gracefully.
// It internally calls `http.Server#Shutdown()`.
func (e *Engine) Shutdown(ctx stdContext.Context) error {
	return e.Server.Shutdown(ctx)
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
