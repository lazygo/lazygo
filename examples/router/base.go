package router

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo/lazygo/examples/app/middleware"
	"github.com/lazygo/lazygo/examples/framework"
)

func Init(app *server.Server) *server.Server {

	// 请求前去除url中 .json 后缀
	app.Pre(middleware.StripUrlSuffix)

	// 拓展middleware后，可在后续的middleware中使用framework.Context
	app.Use(framework.ExtendContextMiddleware)

	// 添加request_id
	app.Use(middleware.RequestID)

	// 添加TrustProxies
	app.Use(middleware.TrustProxies)

	// 增加访问日志记录
	app.Use(middleware.AccessLog)

	app.Get("/", server.NotFoundHandler)

	UserRouter()

	return app
}
