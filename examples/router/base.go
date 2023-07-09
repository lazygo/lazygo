package router

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/examples/framework"
)

func Init(app *server.Server) *server.Server {
	app.Pre(middleware.StripUrlSuffix)
	app.Use(framework.ExtendContextMiddleware)
	app.Use(middleware.AccessLog)
	app.Get("/", server.NotFoundHandler)

	initWechat()

	return app
}
