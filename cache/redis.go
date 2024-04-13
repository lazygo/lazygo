package cache

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/lazygo/lazygo/redis"

	goredis "github.com/redis/go-redis/v9"
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
func (r *redisCache) Remember(key string, fn func() (any, error), ttl int64, ret any) (bool, error) {
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

func (r *redisCache) Set(key string, val any, ttl int64) error {
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

func (r *redisCache) Get(key string, ret any) (bool, error) {
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

func (r *redisCache) HasMulti(keys ...string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	if len(keys) == 0 {
		return result, nil
	}
	ctx := context.Background()

	cmds, err := r.handler.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		for i := range keys {
			pipe.Exists(ctx, r.prefix+keys[i])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		result[keys[i]] = cmd.(*goredis.IntCmd).Val() > 0
	}

	return result, nil
}

func (r *redisCache) Forget(key string) error {
	key = r.prefix + key
	return r.handler.Del(context.Background(), key).Err()
}

func init() {
	registry.Add("redis", newRedisCache)
}
