package main

import (
	"github.com/BurntSushi/toml"
	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/engine"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
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
		Redis []redis.Config `json:"redis" toml:"redis"`
		Cache struct{
			DefaultAdapter string `json:"default" toml:"default"`
			Adapter []cache.Config `json:"adapter" toml:"adapter"`
		} `json:"cache" toml:"cache"`
	}{}

	_, err := toml.DecodeFile("test.toml", &config)
	utils.CheckError(err)
	err = mysql.Init(config.Mysql)
	utils.CheckError(err)
	e.Start("127.0.0.1:1234")
}
