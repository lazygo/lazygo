package framework

import (
	"io"
	"log"
	"os"
	"sync"

	"github.com/lazygo/lazygo/server"
)

var (
	Version        = "0.0.0"
	BuildID        string
	UniqueID       = "F4E8A8E4-0D95-40C3-A560-41BE3CA7C314"
	AppName        = "Lazygo-Examples"
	APPTitle       = AppName
	AppDescription = AppName
)

var (
	httpServer *server.Server
	serverOnce sync.Once
)

func Server() *server.Server {
	serverOnce.Do(func() {
		httpServer = server.New()

		httpServer.HTTPErrorHandler = BaseHTTPErrorHandler(AppHTTPErrorHandler)
		httpServer.HTTPOKHandler = BaseHTTPOKHandler(func(data any, ctx Context) error {
			return ctx.Succ(data)
		})

		// 未经过 lazygo config 初始化日志的场景（如 agent）退化为 stderr，
		// 否则 nil writer 会在 server 首次输出日志时 panic
		var w io.Writer = os.Stderr
		if ErrorLog != nil {
			w = ErrorLog
		}
		httpServer.Logger = log.New(w, "", log.LstdFlags|log.Llongfile)
	})
	return httpServer
}
