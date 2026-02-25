package router

import (
	"github.com/lazygo/lazygo/examples/app/controller"
	"github.com/lazygo/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/server"
)

func AuthRouter(g *server.Group) {
	g = g.Group("", middleware.User)

	// 注册和登陆
	g.Post("register", server.Controller(controller.UserController{}))
	g.Post("forget", server.Controller(controller.UserController{}))
	g.Post("login", server.Controller(controller.UserController{}))
	g.Post("logout", server.Controller(controller.UserController{}))
	g.Post("send_captcha", server.Controller(controller.UserController{}))

	checksign := middleware.CheckSign("vendor", "third_uid", "auth_create")
	g.Post("third", server.Controller(controller.UserController{}, "third_login"), checksign)

	sg := g.Group("", middleware.AuthUser)
	{
		// 此局部函数体内的接口都需要登陆校验
		sg.Post("profile", server.Controller(controller.UserController{}))
		sg.Post("check_login", server.Controller(controller.UserController{}))
		sg.Post("bind_mobile", server.Controller(controller.UserController{}))
	}
}
