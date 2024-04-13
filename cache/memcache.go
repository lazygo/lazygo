package cache

import (
	"encoding/json"
	"errors"
	"reflect"

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

func (m *mcCache) Remember(key string, fn func() (any, error), ttl int64, ret any) (bool, error) {
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
	err = m.handler.Set(key, value, int32(ttl))
	return false, err
}

func (m *mcCache) Set(key string, val any, ttl int64) error {
	key = m.prefix + key
	value, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = m.handler.Set(key, value, int32(ttl))
	if err != nil {
		return err
	}
	return nil
}

func (m *mcCache) Get(key string, ret any) (bool, error) {
	key = m.prefix + key
	item, err := m.handler.Conn().Get(key)
	if err == nil {
		err = json.Unmarshal(item.Value, ret)
		return true, err
	}
	if err != libMemcache.ErrCacheMiss {
		return false, err
	}
	return false, nil
}

func (m *mcCache) Has(key string) (bool, error) {
	key = m.prefix + key
	_, err := m.handler.Conn().Get(key)
	if err == nil {
		return true, nil
	}
	if err != libMemcache.ErrCacheMiss {
		return false, err
	}
	return false, nil
}

func (m *mcCache) HasMulti(keys ...string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	keysWithPrefix := make([]string, len(keys))
	for i := range keys {
		keysWithPrefix[i] = m.prefix + keys[i]
	}
	res, err := m.handler.Conn().GetMulti(keysWithPrefix)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		result[key] = res[keysWithPrefix[i]].Value != nil
	}
	return result, nil
}

func (m *mcCache) Forget(key string) error {
	key = m.prefix + key
	return m.handler.Delete(key)
}

func init() {
	registry.Add("memcache", newMcCache)
}
