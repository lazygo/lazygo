package router

import (
	"github.com/lazygo-dev/lazygo/examples/app/controller"
	"github.com/lazygo/lazygo/server"
)

func InnerRouter(g *server.Group) {
	sg := g.Group("debug")
	{
		sg.Post("dump_post_body", server.Controller(controller.DebugController{}))
		sg.Get("client_ip", server.Controller(controller.DebugController{}, "ClientIP"))
	}

}
