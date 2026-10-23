package cache

import (
	"context"
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
func newRedisCache(opt map[string]string) (backend, error) {
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

func (r *redisCache) Get(key string) ([]byte, bool, error) {
	key = r.prefix + key
	item, err := r.handler.Get(context.Background(), key).Bytes()
	if err == goredis.Nil {
		// 未命中
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return item, true, nil
}

func (r *redisCache) Set(key string, val []byte, ttl int64) error {
	key = r.prefix + key
	return r.handler.Set(context.Background(), key, val, time.Duration(ttl)*time.Second).Err()
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
