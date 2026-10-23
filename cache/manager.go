package cache

import (
	"encoding/json"
	"sync"

	"github.com/lazygo/lazygo/internal"
)

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

// backend 字节级后端接口，T无关，用于适配器注册与Manager存储
type backend interface {
	Get(key string) (data []byte, found bool, err error)
	Set(key string, value []byte, ttl int64) error
	Has(key string) (bool, error)
	HasMulti(keys ...string) (map[string]bool, error)
	Forget(key string) error
}

// Cache 泛型缓存接口，JSON序列化集中在这一层完成
type Cache[T any] interface {
	Get(key string) (T, bool, error)
	Remember(key string, fn func() (T, error), ttl int64) (T, bool, error)
	Set(key string, value T, ttl int64) error
	Has(key string) (bool, error)
	HasMulti(keys ...string) (map[string]bool, error)
	Forget(key string) error
}

type cacheImpl[T any] struct {
	b backend
}

// Get 从缓存获取数据，未命中时返回零值,false,nil
func (c *cacheImpl[T]) Get(key string) (T, bool, error) {
	var ret T
	data, found, err := c.b.Get(key)
	if err != nil || !found {
		return ret, false, err
	}
	err = json.Unmarshal(data, &ret)
	return ret, true, err
}

// Remember 缓存穿透：命中时直接返回缓存值，未命中时调用fn并将结果写入缓存
// 返回值的第二个参数为true表示命中缓存
func (c *cacheImpl[T]) Remember(key string, fn func() (T, error), ttl int64) (T, bool, error) {
	var ret T
	data, found, err := c.b.Get(key)
	if err != nil {
		return ret, false, err
	}
	if found {
		err = json.Unmarshal(data, &ret)
		return ret, true, err
	}
	ret, err = fn()
	if err != nil {
		return ret, false, err
	}
	data, err = json.Marshal(ret)
	if err != nil {
		return ret, false, err
	}
	err = c.b.Set(key, data, ttl)
	return ret, false, err
}

func (c *cacheImpl[T]) Set(key string, value T, ttl int64) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.b.Set(key, data, ttl)
}

func (c *cacheImpl[T]) Has(key string) (bool, error) {
	return c.b.Has(key)
}

func (c *cacheImpl[T]) HasMulti(keys ...string) (map[string]bool, error) {
	return c.b.HasMulti(keys...)
}

func (c *cacheImpl[T]) Forget(key string) error {
	return c.b.Forget(key)
}

type Manager struct {
	sync.Map
	defaultName string
}

var registry = internal.Register[backend, map[string]string]{}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}
		a, err := registry.Get(item.Adapter)
		if err != nil {
			return err
		}
		cache, err := a.Init(item.Option)
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

// Instance 获取缓存实例，每次调用新建泛型外壳（外壳无状态）
func Instance[T any](name string) (Cache[T], error) {
	a, ok := manager.Load(name)
	if !ok {
		return nil, ErrInstanceUninitialized
	}
	return &cacheImpl[T]{b: a.(backend)}, nil
}

// Get 从默认缓存实例获取数据，未命中时ok为false
func Get[T any](key string) (T, bool, error) {
	var zero T
	cache, err := Instance[T](manager.defaultName)
	if err != nil {
		return zero, false, err
	}
	return cache.Get(key)
}

// Remember 缓存穿透：命中时直接返回缓存值，未命中时调用fn并将结果写入缓存
// 返回值的ok为true表示命中缓存
func Remember[T any](key string, fn func() (T, error), ttl int64) (T, bool, error) {
	var zero T
	cache, err := Instance[T](manager.defaultName)
	if err != nil {
		return zero, false, err
	}
	return cache.Remember(key, fn, ttl)
}

func Set(key string, value any, ttl int64) error {
	cache, err := Instance[any](manager.defaultName)
	if err != nil {
		return err
	}
	return cache.Set(key, value, ttl)
}

func Has(key string) (bool, error) {
	cache, err := Instance[any](manager.defaultName)
	if err != nil {
		return false, err
	}
	return cache.Has(key)
}

func HasMulti(keys ...string) (map[string]bool, error) {
	cache, err := Instance[any](manager.defaultName)
	if err != nil {
		return nil, err
	}
	return cache.HasMulti(keys...)
}

func Forget(key string) error {
	cache, err := Instance[any](manager.defaultName)
	if err != nil {
		return err
	}
	return cache.Forget(key)
}
