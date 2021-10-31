package cache

import (
	"encoding/json"
	"github.com/lazygo/lazygo/memcache"
	"time"
)

type mcCache struct {
	name string
	conn *memcache.Memcache
}

// newMcCache 初始化memcache适配器
func newMcCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidMemcacheAdapterParams
	}

	var err error
	conn, err := memcache.Client(name)
	a := &mcCache{
		name: name,
		conn: conn,
	}
	return a, err
}

func (m *mcCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		item, err := m.conn.Conn().Get(key)
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

		err = m.conn.Set(key, value, int32(ttl.Seconds()))
		if err != nil {
			return err
		}
		return nil
	}
	return wp
}

func (m *mcCache) Set(key string, val interface{}, ttl time.Duration) error {
	wp := &wrapper{}
	err := wp.Pack(val, ttl)
	if err != nil {
		return err
	}
	value, err := json.Marshal(wp.Data)
	if err != nil {
		return err
	}
	err = m.conn.Set(key, value, int32(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (m *mcCache) Get(key string) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {

		item, err := m.conn.Conn().Get(key)
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
	wp := &wrapper{}
	item, err := m.conn.Conn().Get(key)
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
	return m.conn.Delete(key)
}

func init() {
	registry.add("memcache", adapterFunc(newMcCache))
}
