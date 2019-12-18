package cache

import (
	"encoding/json"
	"errors"
	"github.com/lazygo/lazygo/memcache"
	"time"
)

type adapterMc struct {
	name      string
	memcached *memcache.Memcache
}

func NewMemcacheAdapter(name string, adapter interface{}) (*adapterMc, error) {
	if mcAdapter, ok := adapter.(*memcache.Memcache); ok {
		a := &adapterMc{
			name,
			mcAdapter,
		}
		return a, nil
	} else {
		return nil, errors.New("Memcache Drive Adapter Failure")
	}
}

func (a adapterMc) Remember(key string, val interface{}, timeout time.Duration) string {
	item, err := a.memcached.Conn().Get(key)
	if err != nil {
		var data string
		if v, ok := val.([]byte); ok {
			data = string(v)
		} else if str, ok := val.(string); ok {
			data = str
		} else {
			panic(errors.New("val only support string and []byte"))
		}
		a.memcached.Set(key, data, int32(timeout.Seconds()))
		//return res
	}
	return string(item.Value)
}

func (a adapterMc) Row(key string, callback func() (map[string]interface{}, error), timeout time.Duration) map[string]interface{} {
	item, err := a.memcached.Conn().Get(key)
	if err != nil {
		res, err := callback()
		if err != nil {
			return nil
		}
		data, err := json.Marshal(res)
		if err != nil {
			return nil
		}
		a.memcached.Set(key, string(data), int32(timeout.Seconds()))
		return res
	}
	var res map[string]interface{}
	json.Unmarshal(item.Value, &res)
	return res
}

func (a adapterMc) List(key string, callback func() ([]map[string]interface{}, error), timeout time.Duration) []map[string]interface{} {
	item, err := a.memcached.Conn().Get(key)
	if err != nil {
		res, err := callback()
		if err != nil {
			return nil
		}
		data, err := json.Marshal(res)
		if err != nil {
			return nil
		}
		a.memcached.Set(key, string(data), int32(timeout.Seconds()))
		return res
	}
	var res []map[string]interface{}
	json.Unmarshal(item.Value, &res)
	return res
}

func (a adapterMc) Has(key string) bool {
	_, err := a.memcached.Conn().Get(key)
	return err != nil
}

func (a adapterMc) Forget(key string) bool {
	err := a.memcached.Delete(key)
	return err == nil
}
