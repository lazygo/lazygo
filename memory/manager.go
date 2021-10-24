package memory

import (
	"sync"
)

type Config struct {
	Name     string `json:"name" toml:"name"`
	Capacity uint64 `json:"capacity" toml:"capacity"`
}

type Manager struct {
	sync.Map
}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}
		m.Store(item.Name, newLRUCache(item.Name, item.Capacity))
	}
	return nil
}

// Init 初始化数据库
func Init(conf []Config) error {
	return manager.init(conf)
}

// LRUCache 通过名称获取LRU实例
func LRUCache(name string) (LRU, error) {
	if lru, ok := manager.Load(name); ok {
		return lru.(LRU), nil
	}
	return nil, ErrLRUCacheNotExists
}
