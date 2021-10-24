package cache

import (
	"time"
)

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

type Cache interface {
	Remember(key string, value func() (interface{}, error), ttl time.Duration) DataResult
	Get(key string) DataResult
	Set(key string, value interface{}, ttl time.Duration) error
	Has(key string) (bool, error)
	Forget(key string) error
}

// Init 初始化设置，在框架初始化时调用
func Init(conf []Config, defaultAdapter string) error {
	return registry.init(conf, defaultAdapter)
}

// Instance 获取缓存实例
func Instance(name string) (Cache, error) {
	a, err := registry.get(name)
	if err != nil {
		return nil, err
	}
	if !a.initialized() {
		return nil, ErrAdapterUninitialized
	}
	return a, nil
}

func Remember(key string, value func() (interface{}, error), ttl time.Duration) DataResult {
	cache, err := Instance(registry.defaultAdapter)
	if err != nil {
		return wrapperError(err)
	}
	return cache.Remember(key, value, ttl)
}

func Get(key string) DataResult {
	cache, err := Instance(registry.defaultAdapter)
	if err != nil {
		return wrapperError(err)
	}
	return cache.Get(key)
}

func Set(key string, value interface{}, ttl time.Duration) error {
	cache, err := Instance(registry.defaultAdapter)
	if err != nil {
		return err
	}
	return cache.Set(key, value, ttl)
}

func Has(key string) (bool, error) {
	cache, err := Instance(registry.defaultAdapter)
	if err != nil {
		return false, err
	}
	return cache.Has(key)
}

func Forget(key string) error {
	cache, err := Instance(registry.defaultAdapter)
	if err != nil {
		return err
	}
	return cache.Forget(key)
}
