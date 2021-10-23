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

// Init 初始化设置
// 在框架初始化时调用
// adapterName 适配器名称
// ext 选项
func Init(conf []*Config) error {
	return registry.init(conf)
}

// Instance 设置缓存
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
