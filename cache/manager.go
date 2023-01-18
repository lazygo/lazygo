package cache

import (
	"sync"
	"time"
)

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

type Cache interface {
	Remember(key string, value func() (interface{}, error), ttl time.Duration, ret interface{}) (bool, error)
	Get(key string, ret interface{}) (bool, error)
	Set(key string, value interface{}, ttl time.Duration) error
	Has(key string) (bool, error)
	Forget(key string) error
}

type Manager struct {
	sync.Map
	defaultName string
}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}
		a, err := registry.get(item.Adapter)
		if err != nil {
			return err
		}
		cache, err := a.init(item.Option)
		if err != nil {
			return err
		}
		m.Store(item.Name, cache)
		if defaultName == item.Name {
			m.defaultName = defaultName
		}
	}
	if m.defaultName == "" {
		return ErrInvalidDefaultName
	}
	return nil
}

// Init 初始化设置，在框架初始化时调用
func Init(conf []Config, defaultAdapter string) error {
	return manager.init(conf, defaultAdapter)
}

// Instance 获取缓存实例
func Instance(name string) (Cache, error) {
	a, ok := manager.Load(name)
	if !ok {
		return nil, ErrAdapterUninitialized
	}
	return a.(Cache), nil
}

func Remember(key string, value func() (interface{}, error), ttl time.Duration, ret interface{}) (bool, error) {
	cache, err := Instance(manager.defaultName)
	if err != nil {
		return false, err
	}
	return cache.Remember(key, value, ttl, ret)
}

func Get(key string, ret interface{}) (bool, error) {
	cache, err := Instance(manager.defaultName)
	if err != nil {
		return false, err
	}
	return cache.Get(key, ret)
}

func Set(key string, value interface{}, ttl time.Duration) error {
	cache, err := Instance(manager.defaultName)
	if err != nil {
		return err
	}
	return cache.Set(key, value, ttl)
}

func Has(key string) (bool, error) {
	cache, err := Instance(manager.defaultName)
	if err != nil {
		return false, err
	}
	return cache.Has(key)
}

func Forget(key string) error {
	cache, err := Instance(manager.defaultName)
	if err != nil {
		return err
	}
	return cache.Forget(key)
}
