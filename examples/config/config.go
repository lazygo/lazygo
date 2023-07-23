package config

import (
	"os"

	"github.com/lazygo/lazygo/examples/pkg/cos"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/config"
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

	// init logger config
	err = base.Register("logger", func(conf Logger) error {
		return logger.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	// init mysql config
	err = base.Register("mysql", mysql.Init)
	if err != nil {
		return err
	}

	// init redis config
	err = base.Register("redis", redis.Init)
	if err != nil {
		return err
	}

	// init memory config
	err = base.Register("memory", memory.Init)
	if err != nil {
		return err
	}

	// init cache config
	err = base.Register("cache", func(conf Cache) error {
		return cache.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	// init locker config
	err = base.Register("locker", func(conf Locker) error {
		return locker.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	// init server config
	err = base.Register("server", func(conf Server) error {
		ServerConfig = conf
		return nil
	})
	if err != nil {
		return err
	}

	// init cos config
	err = base.Register("cos", cos.Init)
	if err != nil {
		return err
	}

	return nil
}
