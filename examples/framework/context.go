package framework

import (
	"strconv"
	"time"

	"github.com/lazygo/lazygo/server"
)

type Context interface {
	server.Context
	Logger() Logger
	GetRequestID() uint64
	GetUID() int64
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
func (c *context) GetRequestID() uint64 {
	if rid, ok := c.Value("request_id").(uint64); ok {
		return rid
	}

	if param := c.GetRequestHeader("request_id"); param != "" {
		rid, err := strconv.ParseUint(param, 10, 64)
		if err != nil {
			c.WithValue("request_id", rid)
			return rid
		}
	}
	rid := GenTraceID()
	c.WithValue("request_id", rid)
	return rid
}

// GetUID 获取请求uid
func (c *context) GetUID() int64 {
	uid, _ := c.Value("uid").(int64)
	return uid
}

// Succ 返回成功
func (c *context) Succ(data interface{}) error {
	result := server.Map{
		"code":  200,
		"errno": 0,
		"msg":   "ok",
		"data":  data,
		"rid":   c.GetRequestID(),
		"t":     time.Now().Unix(),
	}
	return c.JSON(200, result)
}
