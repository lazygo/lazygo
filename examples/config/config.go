package config

import (
	"io"
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

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	base, err := config.Toml(data)
	if err != nil {
		return err
	}
	err = base.Register("cos", cos.Init)
	if err != nil {
		return err
	}

	err = base.Register("logger", func(conf Logger) error {
		return logger.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}

	err = base.Register("mysql", mysql.Init)
	if err != nil {
		return err
	}
	err = base.Register("redis", redis.Init)
	if err != nil {
		return err
	}
	err = base.Register("memory", memory.Init)
	if err != nil {
		return err
	}

	err = base.Register("cache", func(conf Cache) error {
		return cache.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}
	err = base.Register("locker", func(conf Locker) error {
		return locker.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return err
	}
	err = base.Register("server", func(conf Server) error {
		ServerConfig = conf
		return nil
	})
	if err != nil {
		return err
	}
	err = base.Register("cos", cos.Init)
	if err != nil {
		return err
	}

	return nil
}
