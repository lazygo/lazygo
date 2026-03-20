package framework

import (
	"github.com/lazygo/lazygo/server"
)

type Response[T any] struct {
	Code  int    `json:"code"`
	Errno int    `json:"errno"`
	Msg   string `json:"msg"`
	Rid   uint64 `json:"rid"`
	Error string `json:"error,omitempty"`
	Time  int64  `json:"t"`
	Data  T      `json:"data,omitempty"`
}

type HandlerFunc func(Context) error
type HTTPErrorHandler func(error, Context)
type HTTPOKHandler func(any, Context) error

// BaseHandlerFunc HandlerFunc 转为 server.HandlerFunc
func BaseHandlerFunc(h HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc, ok := c.(Context)
		if !ok {
			cc = WrapContext(c)
		}
		return h(cc)
	}
}

// BaseHTTPErrorHandler 返回失败
func BaseHTTPErrorHandler(h HTTPErrorHandler) server.HTTPErrorHandler {
	return func(err error, c server.Context) {
		cc, ok := c.(Context)
		if !ok {
			cc = WrapContext(c)
		}
		h(err, cc)
	}
}

// BaseHTTPOKHandler 返回失败
func BaseHTTPOKHandler(h HTTPOKHandler) server.HTTPOKHandler {
	return func(data any, c server.Context) error {
		cc, ok := c.(Context)
		if !ok {
			cc = WrapContext(c)
		}
		return h(data, cc)
	}
}

// ExtendContextMiddleware Context 拓展中间件
func ExtendContextMiddleware(h server.HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc := WrapContext(c)
		return h(cc)
	}
}

// WrapContext 包装 Context
func WrapContext(c server.Context) Context {
	return &context{c}
}

// HandleSucc 返回成功
func HandleSucc() server.HandlerFunc {
	return BaseHandlerFunc(func(ctx Context) error {
		return ctx.Succ(struct{}{})
	})
}
