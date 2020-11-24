package cache

import (
	"errors"
	"github.com/lazygo/lazygo/redis"
	"time"
)

type redisAdapter struct {
	name string
	conn *redis.Redis
}

// 初始化redis适配器
func (r *redisAdapter) init(opt map[string]interface{}) error {
	if _, ok := opt["name"]; !ok {
		return errors.New("redis适配器参数错误")
	}
	r.name = toString(opt["name"])
	p, err := redis.RedisPool(r.name)
	if err != nil {
		return err
	}
	r.conn = p
	return nil
}

func (r *redisAdapter) Remember(key string, value interface{}, ttl time.Duration) DataResult {
	wp := &wrapper{}
	err := r.conn.GetObject(key, wp)
	if err == nil {
		return wp
	}

	// 穿透
	err = wp.Pack(value)
	CheckError(err)

	err = r.conn.Set(key, wp, int64(ttl.Seconds()))
	CheckError(err)
	//return res
	return wp
}

func (r *redisAdapter) Get(key string) (DataResult, error) {
	wp := &wrapper{}
	err := r.conn.GetObject(key, wp)
	if err == nil {
		return wp, nil
	}
	return nil, err
}

func (r *redisAdapter) Has(key string) bool {
	_, err := r.conn.Get(key)
	return err != nil
}

func (r *redisAdapter) Forget(key string) bool {
	err := r.conn.Del(key)
	return err == nil
}

func init() {
	registry["redis"] = &redisAdapter{}
}
