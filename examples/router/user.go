package router

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo/lazygo/examples/app/controller"
	"github.com/lazygo/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/examples/framework"
)

func UserRouter() {
	app := framework.App()
	g := app.Group("/api/user", middleware.User)
	{

		g.Post("login", server.Controller(controller.UserController{}, "Login"))
		g.Post("logout", server.Controller(controller.UserController{}))

		sg := g.Group("", middleware.AuthUser)
		{
			sg.Post("profile", server.NotFoundHandler)
		}

	}
}
