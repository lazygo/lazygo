package framework

import (
	stdContext "context"
	"fmt"
	"net/http"

	"github.com/lazygo/lazygo/server"
)

type HandlerFunc func(Context) error
type HTTPErrorHandler func(error, Context)
type HTTPOKHandler func(any, Context) error

type stdoutResponseWriter struct{}

func (w *stdoutResponseWriter) Write(p []byte) (n int, err error) {
	fmt.Println(string(p))
	return len(p), nil
}

func (w *stdoutResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *stdoutResponseWriter) WriteHeader(statusCode int) {
	fmt.Println("WriteHeader", statusCode)
}

// BaseHandlerFunc HandlerFunc 转为 server.HandlerFunc
func BaseHandlerFunc(h HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc := c.(Context)
		return h(cc)
	}
}

// BaseHTTPErrorHandler 返回失败
func BaseHTTPErrorHandler(h HTTPErrorHandler) server.HTTPErrorHandler {
	return func(err error, c server.Context) {
		cc := c.(Context)
		h(err, cc)
	}
}

// BaseHTTPOKHandler 返回失败
func BaseHTTPOKHandler(h HTTPOKHandler) server.HTTPOKHandler {
	return func(data any, c server.Context) error {
		cc := c.(Context)
		return h(data, cc)
	}
}

// ExtendContextMiddleware Context 拓展中间件
func ExtendContextMiddleware(h server.HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc := &context{c}
		return h(cc)
	}
}

func NewStdoutContext(ctx stdContext.Context, app *server.Server) Context {
	c := app.NewContext(FakeRequest(ctx), &stdoutResponseWriter{})
	return &context{c}
}

func FakeRequest(ctx stdContext.Context) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "internal://fake-uri", nil)
	return req
}

// HandleSucc 返回成功
func HandleSucc() server.HandlerFunc {
	return BaseHandlerFunc(func(ctx Context) error {
		return ctx.Succ(struct{}{})
	})
}
