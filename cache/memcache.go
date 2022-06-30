package cache

import (
	"encoding/json"
	"github.com/lazygo/lazygo/memcache"
	"time"
)

type mcCache struct {
	name    string
	prefix  string
	handler *memcache.Memcache
}

// newMcCache 初始化memcache适配器
func newMcCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidMemcacheAdapterParams
	}
	prefix := opt["prefix"]

	var err error
	handler, err := memcache.Client(name)
	a := &mcCache{
		name:    name,
		prefix:  prefix,
		handler: handler,
	}
	return a, err
}

func (m *mcCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration) DataResult {
	key = m.prefix + key
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		item, err := m.handler.Conn().Get(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal(item.Value, &wp.Data)
		if err != nil {
			return err
		}

		if wp.Data.Deadline >= time.Now().Unix() {
			return nil
		}

		// 穿透
		err = wp.PackFunc(fn, ttl)
		if err != nil {
			return err
		}

		value, err := json.Marshal(wp.Data)
		if err != nil {
			return err
		}

		return m.handler.Set(key, value, int32(ttl.Seconds()))
	}
	return wp
}

func (m *mcCache) Set(key string, val interface{}, ttl time.Duration) error {
	key = m.prefix + key
	wp := &wrapper{}
	err := wp.Pack(val, ttl)
	if err != nil {
		return err
	}
	value, err := json.Marshal(wp.Data)
	if err != nil {
		return err
	}
	err = m.handler.Set(key, value, int32(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (m *mcCache) Get(key string) DataResult {
	key = m.prefix + key
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {

		item, err := m.handler.Conn().Get(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal(item.Value, &wp.Data)
		if err != nil {
			return err
		}

		if wp.Data.Deadline >= time.Now().Unix() {
			return nil
		}
		return ErrEmptyKey
	}
	return wp
}

func (m *mcCache) Has(key string) (bool, error) {
	key = m.prefix + key
	wp := &wrapper{}
	item, err := m.handler.Conn().Get(key)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(item.Value, &wp.Data)
	if err != nil {
		return false, err
	}
	return wp.Data.Deadline >= time.Now().Unix(), nil
}

func (m *mcCache) Forget(key string) error {
	key = m.prefix + key
	return m.handler.Delete(key)
}

func init() {
	registry.add("memcache", adapterFunc(newMcCache))
}
