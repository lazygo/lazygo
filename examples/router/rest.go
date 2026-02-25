package router

import (
	"github.com/lazygo-dev/lazygo/examples/app/controller"
	"github.com/lazygo-dev/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/server"
)

// RestRouter 注册RESTful API
func RestRouter(g *server.Group) {

	g = g.Group("", middleware.User, middleware.AuthUser)

	g.Get("profile", server.Controller(controller.UserController{}))
	g.Get("connection/:token", server.Controller(controller.CommonController{}))

	sg := g.Group("audit")
	{
		sg.Get("list", server.Controller(controller.AuditController{}))
	}

}
