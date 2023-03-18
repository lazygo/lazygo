package redis

import (
	"fmt"
	"sync"

	goredis "github.com/go-redis/redis/v8"
)

type Config struct {
	Name     string `json:"name" toml:"name"`
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Password string `json:"password" toml:"password"`
	Db       int    `json:"db" toml:"db"`
	Prefix   string `json:"prefix" toml:"prefix"`
}

type Manager struct {
	sync.Map
}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			// 已连接的就不再次连接了
			continue
		}
		client, err := m.open(item)
		if err != nil {
			return err
		}
		m.Store(item.Name, client)
	}
	return nil
}

// 连接redis
func (m *Manager) open(item Config) (*goredis.Client, error) {

	addr := fmt.Sprintf("%s:%d", item.Host, item.Port)

	opts := &goredis.Options{
		Addr:     addr,
		Password: item.Password,
		DB:       item.Db,
	}

	client := goredis.NewClient(opts)
	return client, nil
}

// Init 初始化数据库
func Init(conf []Config) error {
	return manager.init(conf)
}

// Client 通过名称获取Redis
func Client(name string) (*goredis.Client, error) {
	if redis, ok := manager.Load(name); ok {
		return redis.(*goredis.Client), nil
	}
	return nil, ErrRedisNotExists
}
