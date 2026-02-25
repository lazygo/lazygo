package router

import (
	"github.com/lazygo/lazygo/examples/app/controller"
	"github.com/lazygo/lazygo/server"
)

func EventRouter(g *server.Group) {
	g.WebSocket("info", server.Controller(controller.DebugController{}))
	g.Call("info", server.Controller(controller.DebugController{}))
}
