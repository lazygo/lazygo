package cache

import (
	"errors"
	"github.com/lazygo/lazygo/redis"
	"time"
)

type adapterRedis struct {
	name  string
	redis *redis.Redis
}

func NewRedisImpl(name string, adapter interface{}) (*adapterRedis, error) {
	if redisAdapter, ok := adapter.(*redis.Redis); ok {
		a := &adapterRedis{
			name,
			redisAdapter,
		}
		return a, nil
	} else {
		return nil, errors.New("Redis Drive Adapter Failure")
	}
}

func (a adapterRedis) Remember(key string, value interface{}, timeout time.Duration) DataResult {
	wrapper := NewWrapper()
	err := a.redis.GetObject(key, wrapper)
	if err == nil {
		return wrapper
	}

	// 穿透
	err = wrapper.Pack(value)
	CheckError(err)

	err = a.redis.Set(key, wrapper, int64(timeout.Seconds()))
	CheckError(err)
	//return res
	return wrapper
}

func (a adapterRedis) Get(key string) (DataResult, error) {
	wrapper := NewWrapper()
	err := a.redis.GetObject(key, wrapper)
	if err == nil {
		return wrapper, nil
	}
	return nil, err
}

func (a adapterRedis) Has(key string) bool {
	_, err := a.redis.Get(key)
	return err != nil
}

func (a adapterRedis) Forget(key string) bool {
	err := a.redis.Del(key)
	return err == nil
}
