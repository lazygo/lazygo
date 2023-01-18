package cache

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	libMemcache "github.com/bradfitz/gomemcache/memcache"
	"github.com/lazygo/lazygo/memcache"
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

func (m *mcCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration, ret interface{}) (bool, error) {
	key = m.prefix + key
	item, err := m.handler.Conn().Get(key)
	if err == nil {
		err = json.Unmarshal(item.Value, ret)
		return true, err
	}
	if err != libMemcache.ErrCacheMiss {
		return false, err
	}

	// 穿透
	data, err := fn()
	if err != nil {
		return false, err
	}
	rRet := reflect.ValueOf(ret)
	if rRet.Kind() != reflect.Ptr {
		return false, errors.New("ret need a pointer")
	}
	rRet.Elem().Set(reflect.ValueOf(data))

	value, err := json.Marshal(ret)
	if err != nil {
		return false, err
	}
	err = m.handler.Set(key, value, int32(ttl.Seconds()))
	return false, err
}

func (m *mcCache) Set(key string, val interface{}, ttl time.Duration) error {
	key = m.prefix + key
	value, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = m.handler.Set(key, value, int32(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (m *mcCache) Get(key string, ret interface{}) (bool, error) {
	key = m.prefix + key
	item, err := m.handler.Conn().Get(key)
	if err == nil {
		err = json.Unmarshal(item.Value, ret)
		return true, err
	}
	if err != libMemcache.ErrCacheMiss {
		return false, nil
	}

	return false, err
}

func (m *mcCache) Has(key string) (bool, error) {
	key = m.prefix + key
	_, err := m.handler.Conn().Get(key)
	if err != nil {
		if err != libMemcache.ErrCacheMiss {
			return false, err
		}
		return true, nil
	}
	return true, nil
}

func (m *mcCache) Forget(key string) error {
	key = m.prefix + key
	return m.handler.Delete(key)
}

func init() {
	registry.add("memcache", adapterFunc(newMcCache))
}
