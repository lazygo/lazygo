package config

import (
	"errors"
	"os"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/config"
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/pkg/cos"
	"github.com/lazygo/lazygo/locker"
	"github.com/lazygo/lazygo/logger"
	"github.com/lazygo/lazygo/memory"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
)

type Logger struct {
	DefaultName string          `json:"default" toml:"default"`
	Adapter     []logger.Config `json:"adapter" toml:"adapter"`
}

type Cache struct {
	DefaultName string         `json:"default" toml:"default"`
	Adapter     []cache.Config `json:"adapter" toml:"adapter"`
}
type Locker struct {
	DefaultName string          `json:"default" toml:"default"`
	Adapter     []locker.Config `json:"adapter" toml:"adapter"`
}

type Server struct {
	Addr  string `json:"addr" toml:"addr"`
	Debug bool   `json:"debug" toml:"debug"`
}

var ServerConfig Server

func Init(filename string) error {

	// load config file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var base *config.Config
	for _, loader := range []config.Loader{config.Json, config.Toml} {
		base, err = loader(data)
		if err == nil {
			break
		}
	}
	if err != nil || base == nil {
		return errors.New("config file format fail")
	}

	// load logger config
	err = base.Load("logger", func(conf Logger) error {
		return logger.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}
	framework.InitLogger()

	// load mysql config
	err = base.Load("mysql", mysql.Init)
	if err != nil {
		return err
	}

	// load redis config
	err = base.Load("redis", redis.Init)
	if err != nil {
		return err
	}

	// load memory config
	err = base.Load("memory", memory.Init)
	if err != nil {
		return err
	}

	// load cache config
	err = base.Load("cache", func(conf Cache) error {
		return cache.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	// load locker config
	err = base.Load("locker", func(conf Locker) error {
		return locker.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	// load server config
	err = base.Load("server", func(conf Server) error {
		ServerConfig = conf
		return nil
	})
	if err != nil {
		return err
	}

	// load cos config
	err = base.Load("cos", cos.Init)
	if err != nil {
		return err
	}

	return nil
}
