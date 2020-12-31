package main

import (
	"fmt"
	"github.com/lazygo/lazygo/config"
	"github.com/lazygo/lazygo/engine"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/test/app"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	e := engine.New()
	e.Get("/ab/:xxx/cd.json", app.TestController{}.TestResponseAction)
	config.LoadFile("test")
	conf, err := config.GetSection("mysql")
	if err != nil {
		fmt.Println(err)
	}
	mysql.Init(conf)
	e.Start("127.0.0.1:1234")
}
