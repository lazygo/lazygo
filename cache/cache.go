package cache

import (
	"errors"
	"github.com/tidwall/gjson"
	"time"
)

type DataResult interface {
	GetType() int32
	ToString() (string, error)
	ToByteArray() ([]byte, error)
	ToMapIf() (MapIf, error)
	ToMapIfArray() (MapIfArray, error)
}

type Cache interface {
	Remember(key string, value interface{}, timeout time.Duration) DataResult
	Get(key string) (DataResult, error)
	Has(key string) bool
	Forget(key string) bool
}

/*
{"driver": "redis", "name": "redis1"}
*/

func NewCache(conf *gjson.Result, getAdapter func(driver string, name string) interface{}) (Cache, error) {
	driver := conf.Get("driver").String()
	name := conf.Get("name").String()

	if driver == "memcache" {
		adapter := getAdapter("memcache", name)
		return NewMemcacheImpl(name, adapter)
	}

	if driver == "redis" {
		adapter := getAdapter("redis", name)
		return NewRedisImpl(name, adapter)
	}
	return nil, errors.New("不支持此驱动: " + driver)
}
