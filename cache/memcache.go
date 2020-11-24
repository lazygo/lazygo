package cache

import (
	"encoding/json"
	"errors"
	"github.com/lazygo/lazygo/memcache"
	"time"
)

type mcAdapter struct {
	name string
	conn *memcache.Memcache
}

func (m *mcAdapter) init(opt map[string]interface{}) error {
	if _, ok := opt["name"]; !ok {
		return errors.New("memcached适配器参数错误")
	}
	m.name = toString(opt["name"])
	p, err := memcache.Mc(m.name)
	if err != nil {
		return err
	}
	m.conn = p
	return nil
}

func (m *mcAdapter) Remember(key string, value interface{}, ttl time.Duration) DataResult {
	item, err := m.conn.Conn().Get(key)
	wrapper := &wrapper{}

	if err == nil {
		err := json.Unmarshal(item.Value, wrapper)
		if err == nil {
			return wrapper
		}
	}

	// 穿透
	err = wrapper.Pack(value)
	CheckError(err)

	data, err := json.Marshal(wrapper)
	CheckError(err)

	m.conn.Set(key, data, int32(ttl.Seconds()))
	//return res
	return wrapper
}

func (m *mcAdapter) Get(key string) (DataResult, error) {
	item, err := m.conn.Conn().Get(key)
	wrapper := &wrapper{}

	if err == nil {
		err := json.Unmarshal(item.Value, wrapper)
		if err == nil {
			return wrapper, nil
		}
	}
	return nil, err
}

func (m *mcAdapter) Has(key string) bool {
	_, err := m.conn.Conn().Get(key)
	return err != nil
}

func (m *mcAdapter) Forget(key string) bool {
	err := m.conn.Delete(key)
	return err == nil
}

func init() {
	registry["memcached"] = &mcAdapter{}
}
