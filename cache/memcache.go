package cache

import (
	libMemcache "github.com/bradfitz/gomemcache/memcache"
	"github.com/lazygo/lazygo/memcache"
)

type mcCache struct {
	name    string
	prefix  string
	handler *memcache.Memcache
}

// newMcCache 初始化memcache适配器
func newMcCache(opt map[string]string) (backend, error) {
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

func (m *mcCache) Get(key string) ([]byte, bool, error) {
	key = m.prefix + key
	item, err := m.handler.Conn().Get(key)
	if err == libMemcache.ErrCacheMiss {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return item.Value, true, nil
}

func (m *mcCache) Set(key string, val []byte, ttl int64) error {
	key = m.prefix + key
	return m.handler.Set(key, val, int32(ttl))
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
		// 未命中的key在res中不存在对应item，需判空
		if item, ok := res[keysWithPrefix[i]]; ok && item != nil {
			result[key] = true
		} else {
			result[key] = false
		}
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
