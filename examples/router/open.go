package router

import (
	"github.com/lazygo-dev/lazygo/examples/app/controller"
	"github.com/lazygo-dev/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/server"
)

// OpenRouter 开放平台
func OpenRouter(g *server.Group) {

	// 注册设备
	g = g.Group("/v1", middleware.Third, middleware.AuthUser)
	g.Post("/regist_device", server.Controller(controller.OpenColtroller{}))
}
