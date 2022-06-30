package cache

import (
	"encoding/json"
	"github.com/lazygo/lazygo/memory"
	"time"
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

func (l *lruCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration) DataResult {
	key = l.prefix + key
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		item, ok := l.handler.Get(key)
		if ok {
			err := json.Unmarshal(item, &wp.Data)
			if err != nil {
				return err
			}
			if wp.Data.Deadline >= time.Now().Unix() {
				return nil
			}
		}

		// 穿透
		err := wp.PackFunc(fn, ttl)
		if err != nil {
			return err
		}

		value, err := json.Marshal(wp.Data)
		if err != nil {
			return err
		}

		return l.handler.Set(key, value, int32(ttl.Seconds()))
	}
	return wp
}

func (l *lruCache) Set(key string, val interface{}, ttl time.Duration) error {
	key = l.prefix + key
	wp := &wrapper{}
	err := wp.Pack(val, ttl)
	if err != nil {
		return err
	}
	value, err := json.Marshal(wp.Data)
	if err != nil {
		return err
	}
	err = l.handler.Set(key, value, int32(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (l *lruCache) Get(key string) DataResult {
	key = l.prefix + key
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		if item, ok := l.handler.Get(key); ok {
			err := json.Unmarshal(item, &wp.Data)
			if err != nil {
				return err
			}

			if wp.Data.Deadline >= time.Now().Unix() {
				return nil
			}
		}
		return ErrEmptyKey
	}
	return wp
}

func (l *lruCache) Has(key string) (bool, error) {
	key = l.prefix + key
	wp := &wrapper{}

	if item, ok := l.handler.Get(key); ok {
		err := json.Unmarshal(item, &wp.Data)
		if err != nil {
			return false, err
		}

		return wp.Data.Deadline >= time.Now().Unix(), nil
	}
	return false, nil
}

func (l *lruCache) Forget(key string) error {
	key = l.prefix + key
	l.handler.Delete(key)
	return nil
}

func init() {
	registry.add("lru", adapterFunc(newLruCache))
}
