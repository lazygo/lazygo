package cache

import (
	redigo "github.com/gomodule/redigo/redis"
	"github.com/lazygo/lazygo/redis"
	"time"
)

type redisCache struct {
	name    string
	handler *redis.Redis
}

// newRedisCache 初始化redis适配器
func newRedisCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidRedisAdapterParams
	}

	var err error
	handler, err := redis.Pool(name)
	a := &redisCache{
		name:    name,
		handler: handler,
	}
	return a, err
}

func (r *redisCache) Remember(key string, fn func() (interface{}, error), ttl time.Duration) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		err := r.handler.GetObject(key, &wp.Data)
		if err != nil && err != redigo.ErrNil {
			return err
		}

		if err != redigo.ErrNil && wp.Data.Deadline >= time.Now().Unix() {
			return nil
		}

		// 穿透
		err = wp.PackFunc(fn, ttl)
		if err != nil {
			return err
		}

		return r.handler.Set(key, wp.Data, int64(ttl.Seconds()))
	}
	return wp
}

func (r *redisCache) Set(key string, val interface{}, ttl time.Duration) error {
	wp := &wrapper{}
	err := wp.Pack(val, ttl)
	if err != nil {
		return err
	}
	err = r.handler.Set(key, wp, int64(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Get(key string) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		err := r.handler.GetObject(key, &wp.Data)
		if err != nil && err != redigo.ErrNil {
			return err
		}

		if err != redigo.ErrNil && wp.Data.Deadline >= time.Now().Unix() {
			return nil
		}

		return ErrEmptyKey
	}
	return wp
}

func (r *redisCache) Has(key string) (bool, error) {
	wp := &wrapper{}
	err := r.handler.GetObject(key, &wp.Data)
	if err == redigo.ErrNil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return wp.Data.Deadline >= time.Now().Unix(), nil
}

func (r *redisCache) Forget(key string) error {
	return r.handler.Del(key)
}

func init() {
	registry.add("redis", adapterFunc(newRedisCache))
}
