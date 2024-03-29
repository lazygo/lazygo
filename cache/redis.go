package cache

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/lazygo/lazygo/redis"

	goredis "github.com/go-redis/redis/v8"
)

type redisCache struct {
	name    string
	prefix  string
	handler *goredis.Client
}

// newRedisCache 初始化redis适配器
func newRedisCache(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidRedisAdapterParams
	}
	prefix := opt["prefix"]

	var err error
	handler, err := redis.Client(name)
	a := &redisCache{
		name:    name,
		prefix:  prefix,
		handler: handler,
	}
	return a, err
}

// Remember 获取缓存，如果没有命中缓存则使用fn实时获取
func (r *redisCache) Remember(key string, fn func() (interface{}, error), ttl int64, ret interface{}) (bool, error) {
	key = r.prefix + key
	item, err := r.handler.Get(context.Background(), key).Bytes()
	if err == nil {
		err = json.Unmarshal(item, ret)
		return true, err
	}
	if err != goredis.Nil {
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

	value, err := json.Marshal(ret)
	if err != nil {
		return false, err
	}
	err = r.handler.Set(context.Background(), key, value, time.Duration(ttl)*time.Second).Err()

	return false, err
}

func (r *redisCache) Set(key string, val interface{}, ttl int64) error {
	key = r.prefix + key
	value, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = r.handler.Set(context.Background(), key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *redisCache) Get(key string, ret interface{}) (bool, error) {
	key = r.prefix + key
	item, err := r.handler.Get(context.Background(), key).Bytes()
	if err == nil {
		err = json.Unmarshal(item, ret)
		return true, err
	}
	if err != goredis.Nil {
		return false, err
	}

	return true, nil
}

func (r *redisCache) Has(key string) (bool, error) {
	key = r.prefix + key
	n, err := r.handler.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *redisCache) Forget(key string) error {
	key = r.prefix + key
	return r.handler.Del(context.Background(), key).Err()
}

func init() {
	registry.add("redis", adapterFunc(newRedisCache))
}
