package main

import (
	"github.com/lazygo/lazygo/engine"
	"github.com/lazygo/lazygo/test/app"
)

func main() {
	e := engine.New()
	e.Get("/ab/cd", app.TestController{}.TestResponseAction)
	e.Start("127.0.0.1:1234")
}
