package cache

import (
	"github.com/lazygo/lazygo/memory"
)

type lruCache struct {
	name    string
	prefix  string
	handler memory.LRU
}

// newLruCache 初始化LRU适配器
func newLruCache(opt map[string]string) (backend, error) {
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

func (l *lruCache) Get(key string) ([]byte, bool, error) {
	key = l.prefix + key
	item, ok := l.handler.Get(key)
	if !ok {
		return nil, false, nil
	}
	return item, true, nil
}

func (l *lruCache) Set(key string, val []byte, ttl int64) error {
	key = l.prefix + key
	return l.handler.Set(key, val, int32(ttl))
}

func (l *lruCache) Has(key string) (bool, error) {
	key = l.prefix + key
	return l.handler.Exists(key), nil
}

func (l *lruCache) HasMulti(keys ...string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	for i := range keys {
		result[keys[i]] = l.handler.Exists(l.prefix + keys[i])
	}
	return result, nil
}

func (l *lruCache) Forget(key string) error {
	key = l.prefix + key
	l.handler.Delete(key)
	return nil
}

func init() {
	registry.Add("memory", newLruCache)
}
