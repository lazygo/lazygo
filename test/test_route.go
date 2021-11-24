package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/locker"
	"github.com/lazygo/lazygo/memcache"
	"github.com/lazygo/lazygo/memory"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/lazygo/test/app"
	"github.com/lazygo/lazygo/utils"
	"runtime"
)

type Config struct {
	Mysql    []mysql.Config    `json:"mysql" toml:"mysql"`
	Redis    []redis.Config    `json:"redis" toml:"redis"`
	Memcache []memcache.Config `json:"memcache" toml:"memcache"`
	Memory   []memory.Config   `json:"memory" toml:"memory"`
	Cache    struct {
		DefaultName string         `json:"default" toml:"default"`
		Adapter     []cache.Config `json:"adapter" toml:"adapter"`
	} `json:"cache" toml:"cache"`
	Locker struct {
		DefaultName string          `json:"default" toml:"default"`
		Adapter     []locker.Config `json:"adapter" toml:"adapter"`
	} `json:"locker" toml:"locker"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	e := server.New()
	e.Get("/ab/:xxx/cd.json", app.TestController{}.TestResponseAction)

	config := Config{}

	_, err := toml.DecodeFile("test.toml", &config)
	utils.CheckError(err)

	v, _ := json.Marshal(config)
	fmt.Println(string(v))

	err = mysql.Init(config.Mysql)
	utils.CheckError(err)

	err = redis.Init(config.Redis)
	utils.CheckError(err)

	err = memcache.Init(config.Memcache)
	utils.CheckError(err)

	err = memory.Init(config.Memory)
	utils.CheckError(err)

	err = locker.Init(config.Locker.Adapter, config.Locker.DefaultName)
	utils.CheckError(err)

	err = cache.Init(config.Cache.Adapter, config.Cache.DefaultName)
	utils.CheckError(err)

	e.Start("127.0.0.1:1234")
}
