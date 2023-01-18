package cache

import (
	"errors"
	"reflect"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/lazygo/lazygo/redis"
)

type redisCache struct {
	name    string
	prefix  string
	handler *redis.Redis
}

// newRedisCache 初始化redis适配器
func newRedisCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidRedisAdapterParams
	}
	prefix := opt["prefix"]

	var err error
	handler, err := redis.Pool(name)
	a := &redisCache{
		name:    name,
		prefix:  prefix,
		handler: handler,
	}
	return a, err
}

func (r *redisCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration, ret interface{}) (bool, error) {
	key = r.prefix + key
	err := r.handler.GetObject(key, ret)
	if err == nil {
		return true, nil
	}
	if err != redigo.ErrNil {
		return false, err
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

	err = r.handler.Set(key, data, int64(ttl.Seconds()))

	return false, err
}

func (r *redisCache) Set(key string, val interface{}, ttl time.Duration) error {
	key = r.prefix + key
	err := r.handler.Set(key, val, int64(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Get(key string, ret interface{}) (bool, error) {
	key = r.prefix + key
	err := r.handler.GetObject(key, ret)
	if err != nil && err != redigo.ErrNil {
		return false, err
	}

	return true, nil
}

func (r *redisCache) Has(key string) (bool, error) {
	key = r.prefix + key
	ok, err := r.handler.Exists(key)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (r *redisCache) Forget(key string) error {
	key = r.prefix + key
	return r.handler.Del(key)
}

func init() {
	registry.add("redis", adapterFunc(newRedisCache))
}
