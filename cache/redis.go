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

func NewRedisAdapter(name string, adapter interface{}) (*adapterRedis, error) {
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

func (a adapterRedis) Remember(key string, val interface{}, timeout time.Duration) string {
	item, err := a.redis.GetString(key)
	if err != nil {
		var data string
		if v, ok := val.([]byte); ok {
			data = string(v)
		} else if str, ok := val.(string); ok {
			data = str
		} else {
			panic(errors.New("val only support string and []byte"))
		}
		a.redis.Set(key, data, int64(timeout.Seconds()))
		//return res
	}
	return item
}

func (a adapterRedis) Row(key string, callback func() (map[string]interface{}, error), timeout time.Duration) map[string]interface{} {
	item, err := a.redis.Get(key)
	if err != nil {
		res, err := callback()
		if err != nil {
			return nil
		}
		a.redis.Set(key, res, int64(timeout.Seconds()))
		return res
	}
	return item.(map[string]interface{})
}

func (a adapterRedis) List(key string, callback func() ([]map[string]interface{}, error), timeout time.Duration) []map[string]interface{} {
	item, err := a.redis.Get(key)
	if err != nil {
		res, err := callback()
		if err != nil {
			return nil
		}
		a.redis.Set(key, res, int64(timeout.Seconds()))
		return res
	}
	return item.([]map[string]interface{})
}

func (a adapterRedis) Has(key string) bool {
	_, err := a.redis.Get(key)
	return err != nil
}

func (a adapterRedis) Forget(key string) bool {
	err := a.redis.Del(key)
	return err == nil
}
