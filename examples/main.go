package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/lazygo/lazygo/examples/config"
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/router"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	err := config.Init("./config.toml")
	if err != nil {
		log.Fatalf("[msg: load config file error] [err: %v]", err)
	}

	app := framework.App()
	app.Debug = config.ServerConfig.Debug

	fmt.Println("Listen " + config.ServerConfig.Addr)
	err = router.Init(app).Start(config.ServerConfig.Addr)
	if err != nil {
		log.Fatalln(err)
	}
}
