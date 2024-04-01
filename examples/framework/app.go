package framework

import (
	"log"
	"os"

	"github.com/lazygo/lazygo/server"
)

var app *server.Server

var Hostname string

func App() *server.Server {
	if app != nil {
		return app
	}
	app = server.New()

	app.HTTPErrorHandler = BaseHTTPErrorHandler(AppHTTPErrorHandler)
	app.HTTPOKHandler = BaseHTTPOKHandler(func(data interface{}, ctx Context) error {
		return ctx.Succ(data)
	})
	InitLogger()
	app.Logger = log.New(ErrorLog, "", log.LstdFlags&log.Llongfile)

	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		log.Fatalf("[msg: get hostname fail] [err: %v]", err)
	}

	return app
}
