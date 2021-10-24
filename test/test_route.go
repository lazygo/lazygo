package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lazygo/lazygo/engine"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/test/app"
	"github.com/lazygo/lazygo/utils"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	e := engine.New()
	e.Get("/ab/:xxx/cd.json", app.TestController{}.TestResponseAction)

	config := struct {
		Mysql []mysql.Config `json:"mysql" toml:"mysql"`
	}{}

	_, err := toml.DecodeFile("test.toml", &config)
	utils.CheckError(err)
	fmt.Println(config.Mysql[0])
	err = mysql.Init(config.Mysql)
	utils.CheckError(err)
	e.Start("127.0.0.1:1234")
}
