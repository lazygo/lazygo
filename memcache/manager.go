package memcache

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"sync"
	"time"
)

type ServerConfig struct {
	Host string `json:"host" toml:"host"`
	Port int    `json:"port" toml:"port"`
}

type Config struct {
	Name   string         `json:"name" toml:"name"`
	Server []ServerConfig `json:"server" toml:"server"`
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
		mc, err := m.open(item)
		if err != nil {
			return err
		}
		m.Store(item.Name, newMemcache(item.Name, mc))
	}
	return nil
}

// 连接memcache
func (m *Manager) open(item Config) (*memcache.Client, error) {
	var serverList []string
	for _, server := range item.Server {
		serverList = append(serverList, fmt.Sprintf("%s:%d", server.Host, server.Port))
	}
	mc := memcache.New(serverList...)
	mc.MaxIdleConns = 10 // 最大保持10个空闲连接
	mc.Timeout = time.Duration(10) * time.Second
	return mc, nil
}

// Init 初始化数据库
func Init(conf []Config) error {
	return manager.init(conf)
}

// Client 通过名称获取Memcache
func Client(name string) (*Memcache, error) {
	if mc, ok := manager.Load(name); ok {
		return mc.(*Memcache), nil
	}
	return nil, ErrMemcacheNotExists
}
