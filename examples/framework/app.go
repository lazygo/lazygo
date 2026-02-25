package framework

import (
	"log"
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

		httpServer.Logger = log.New(ErrorLog, "", log.LstdFlags&log.Llongfile)
	})
	return httpServer
}
