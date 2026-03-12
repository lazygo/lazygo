package framework

import (
	"net"
	"net/http"
	"strings"
	"time"

	stdContext "context"

	"github.com/lazygo/lazygo/server"
)

var (
	_ Context            = (*context)(nil)
	_ stdContext.Context = (*context)(nil)
)

type Context interface {
	server.Context
	Logger() Logger
	RequestID() uint64
	UID() uint64
	RealIP() string
	Succ(any) error
}

type context struct {
	server.Context
}

// Logger 日志记录器
func (c *context) Logger() Logger {
	return &loggerImpl{
		Ctx: c,
	}
}

// RequestID 获取请求id
func (c *context) RequestID() uint64 {
	rid, _ := c.Value(server.HeaderXRequestID).(uint64)
	return rid
}

// UID 获取请求uid
func (c *context) UID() uint64 {
	uid, _ := c.Value("uid").(uint64)
	return uid
}

// RealIP 获取请求客户端IP
func (c *context) RealIP() string {
	if realIP, ok := c.Value(server.HeaderXRealIP).(string); ok {
		return realIP
	}
	// Fall back to legacy behavior
	if ip := c.RequestHeader(server.HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			xffip := strings.TrimSpace(ip[:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return ip
	}
	if ip := c.RequestHeader(server.HeaderXRealIP); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
	return ra
}

// Succ 返回成功
func (c *context) Succ(data any) error {
	resp := Response[any]{
		Code: http.StatusOK,
		Msg:  http.StatusText(http.StatusOK),
		Data: data,
		Time: time.Now().Unix(),
	}
	rid := c.RequestID()
	if rid != 0 {
		resp.Rid = rid
	}
	return c.JSON(http.StatusOK, resp)
}
