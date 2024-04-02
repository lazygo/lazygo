package framework

import (
	"github.com/lazygo/lazygo/server"
)

type HandlerFunc func(Context) error
type HTTPErrorHandler func(error, Context)
type HTTPOKHandler func(interface{}, Context) error

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
	return func(data interface{}, c server.Context) error {
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

// HandleSucc 返回成功
func HandleSucc() server.HandlerFunc {
	return BaseHandlerFunc(func(ctx Context) error {
		return ctx.Succ(struct{}{})
	})
}
