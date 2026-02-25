package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/config"
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	"github.com/lazygo/lazygo/httpclient/httpdns"
	"github.com/lazygo/lazygo/locker"
	"github.com/lazygo/lazygo/logger"
	"github.com/lazygo/lazygo/memory"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
	"github.com/lazygo/pkg/cos"
	"github.com/lazygo/pkg/mail"
	"github.com/lazygo/pkg/sms"
	"github.com/lazygo/pkg/wechat"
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

type HTTPDNS struct {
	DefaultName string           `json:"default" toml:"default"`
	Adapter     []httpdns.Config `json:"adapter" toml:"adapter"`
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
		return fmt.Errorf("read config file fail: %w", err)
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
		return fmt.Errorf("init logger fail: %w", err)
	}
	framework.InitLogger()

	// load mysql config
	err = base.Load("mysql", mysql.Init)
	if err != nil {
		return fmt.Errorf("init mysql fail: %w", err)
	}

	// load redis config
	err = base.Load("redis", redis.Init)
	if err != nil {
		return fmt.Errorf("init redis fail: %w", err)
	}

	// load memory config
	err = base.Load("memory", memory.Init)
	if err != nil {
		return fmt.Errorf("init memory fail: %w", err)
	}

	// load cache config
	err = base.Load("cache", func(conf Cache) error {
		return cache.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return fmt.Errorf("init cache fail: %w", err)
	}

	// load locker config
	err = base.Load("locker", func(conf Locker) error {
		return locker.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return fmt.Errorf("init locker fail: %w", err)
	}

	// load httpdns
	err = base.Load("httpdns", func(conf HTTPDNS) error {
		return httpdns.Init(conf.Adapter, conf.DefaultName)
	})
	if err != nil {
		return fmt.Errorf("init httpdns fail: %w", err)
	}

	// load cos config
	err = base.Load("cos", func(conf cos.Config) error {
		return cos.Init(conf)
	})
	if err != nil {
		return fmt.Errorf("init cos fail: %w", err)
	}

	// load smtp config
	err = base.Load("smtp", func(conf mail.SmtpConfig) error {
		return mail.Init(conf)
	})
	if err != nil {
		return fmt.Errorf("init smtp fail: %w", err)
	}

	// load sms config
	err = base.Load("sms", func(conf sms.TencentConfig) error {
		return sms.Init(conf)
	})
	if err != nil {
		return fmt.Errorf("init sms fail: %w", err)
	}

	// load wechat_official_account config
	err = base.Load("wechat_official_account", func(conf wechat.OfficialAccountConfig) error {
		return wechat.InitOfficialAccount(conf, cacheModel.NewWechatCache())
	})
	if err != nil {
		return fmt.Errorf("init wechat official account fail: %w", err)
	}

	// load server config
	err = base.Load("server", func(conf Server) error {
		ServerConfig = conf
		return nil
	})
	if err != nil {
		return fmt.Errorf("init server fail: %w", err)
	}

	return nil
}
