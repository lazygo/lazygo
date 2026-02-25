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
	app.Pre(framework.ExtendContextMiddleware)

	// 添加request_id
	app.Use(middleware.RequestID)

	// 添加TrustProxies
	app.Use(middleware.TrustProxies)

	// 支持解压body
	app.Use(middleware.DecompressRequest)

	// 增加访问日志记录
	app.Use(middleware.AccessLog)

	// Debug 模式开启跨域支持
	if app.Debug {
		app.Use(middleware.Debug)
	}

	app.Get("/", server.NotFoundHandler)

	InnerRouter(app.Group("/internal"))
	RestRouter(app.Group("/api/rest"))
	AuthRouter(app.Group("/api/auth"))
	ThirdRouter(app.Group("/api/third"))
	OpenRouter(app.Group("/api/open"))
	EventRouter(app.Group("/event"))

	return app
}
