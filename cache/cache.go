package cache

import (
	"errors"
	"github.com/tidwall/gjson"
	"time"
)

type ICache interface {
	Remember(key string, val interface{}, timeout time.Duration) string
	Row(key string, callback func() (map[string]interface{}, error), timeout time.Duration) map[string]interface{}
	List(key string, callback func() ([]map[string]interface{}, error), timeout time.Duration) []map[string]interface{}
	Has(key string) bool
	Forget(key string) bool
}

/*
{"driver": "redis", "name": "redis1"}
*/

func NewCache(conf *gjson.Result, getAdapter func(driver string, name string) interface{}) (ICache, error) {
	driver := conf.Get("driver").String()
	name := conf.Get("name").String()

	if driver == "memcache" {
		adapter := getAdapter("memcache", name)
		return NewMemcacheAdapter(name, adapter)
	}

	if driver == "redis" {
		adapter := getAdapter("redis", name)
		return NewRedisAdapter(name, adapter)
	}
	return nil, errors.New("不支持此驱动: " + driver)
}
