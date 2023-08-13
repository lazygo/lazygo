package framework

import (
	"net"
	"strings"
	"time"

	"github.com/lazygo/lazygo/server"
)

type Context interface {
	server.Context
	Logger() Logger
	RequestID() uint64
	UID() int64
	RealIP() string
	Succ(interface{}) error
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

// GetRequestID 获取请求id
func (c *context) RequestID() uint64 {
	rid, _ := c.Value(server.HeaderXRequestID).(uint64)
	return rid
}

// GetUID 获取请求uid
func (c *context) UID() int64 {
	uid, _ := c.Value("uid").(int64)
	return uid
}

// GetUID 获取请求uid
func (c *context) RealIP() string {
	if realIP, ok := c.Value("real_ip").(string); ok {
		return realIP
	}
	// Fall back to legacy behavior
	if ip := c.GetRequestHeader(server.HeaderXForwardedFor); ip != "" {
		i := strings.IndexAny(ip, ",")
		if i > 0 {
			xffip := strings.TrimSpace(ip[:i])
			xffip = strings.TrimPrefix(xffip, "[")
			xffip = strings.TrimSuffix(xffip, "]")
			return xffip
		}
		return ip
	}
	if ip := c.GetRequestHeader(server.HeaderXRealIP); ip != "" {
		ip = strings.TrimPrefix(ip, "[")
		ip = strings.TrimSuffix(ip, "]")
		return ip
	}
	ra, _, _ := net.SplitHostPort(c.Request().RemoteAddr)
	return ra
}

// Succ 返回成功
func (c *context) Succ(data interface{}) error {
	result := server.Map{
		"code":  200,
		"errno": 0,
		"msg":   "ok",
		"data":  data,
		"rid":   c.RequestID(),
		"t":     time.Now().Unix(),
	}
	rid := c.RequestID()
	if rid != 0 {
		result["rid"] = rid
	}
	return c.JSON(200, result)
}
