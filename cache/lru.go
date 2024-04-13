package cache

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/lazygo/lazygo/memory"
)

type lruCache struct {
	name    string
	prefix  string
	handler memory.LRU
}

// newLruCache 初始化LRU适配器
func newLruCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidLruAdapterParams
	}
	prefix := opt["prefix"]

	var err error
	lru, err := memory.LRUCache(name)
	l := &lruCache{
		name:    name,
		prefix:  prefix,
		handler: lru,
	}
	return l, err
}

func (l *lruCache) Remember(key string, fn func() (any, error), ttl int64, ret any) (bool, error) {
	key = l.prefix + key
	if item, ok := l.handler.Get(key); ok {
		err := json.Unmarshal(item, ret)
		return true, err
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
	err = l.handler.Set(key, value, int32(ttl))

	return false, err
}

func (l *lruCache) Set(key string, val any, ttl int64) error {
	key = l.prefix + key
	value, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = l.handler.Set(key, value, int32(ttl))
	if err != nil {
		return err
	}
	return nil
}

func (l *lruCache) Get(key string, ret any) (bool, error) {
	key = l.prefix + key
	if item, ok := l.handler.Get(key); ok {
		err := json.Unmarshal(item, ret)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (l *lruCache) Has(key string) (bool, error) {
	key = l.prefix + key

	if l.handler.Exists(key) {
		return true, nil
	}
	return false, nil
}

func (l *lruCache) HasMulti(keys ...string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	for i := range keys {
		ok, err := l.Has(l.prefix + keys[i])
		if err != nil {
			return nil, err
		}
		result[keys[i]] = ok
	}
	return result, nil
}

func (l *lruCache) Forget(key string) error {
	key = l.prefix + key
	l.handler.Delete(key)
	return nil
}

func init() {
	registry.Add("lru", newLruCache)
}
