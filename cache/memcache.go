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

func NewMemcacheImpl(name string, adapter interface{}) (*adapterMc, error) {
	if mcAdapter, ok := adapter.(*memcache.Memcache); ok {
		a := &adapterMc{
			name,
			mcAdapter,
		}
		return a, nil
	}
	return nil, errors.New("Memcache Drive Adapter Failure")
}

func (a adapterMc) Remember(key string, value interface{}, timeout time.Duration) DataResult {
	item, err := a.memcached.Conn().Get(key)
	wrapper := NewWrapper()

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

	a.memcached.Set(key, data, int32(timeout.Seconds()))
	//return res
	return wrapper
}

func (a adapterMc) Get(key string) (DataResult, error) {
	item, err := a.memcached.Conn().Get(key)
	wrapper := NewWrapper()

	if err == nil {
		err := json.Unmarshal(item.Value, wrapper)
		if err == nil {
			return wrapper, nil
		}
	}
	return nil, err
}

func (a adapterMc) Has(key string) bool {
	_, err := a.memcached.Conn().Get(key)
	return err != nil
}

func (a adapterMc) Forget(key string) bool {
	err := a.memcached.Delete(key)
	return err == nil
}
