package cache

import (
	redigo "github.com/gomodule/redigo/redis"
	"github.com/lazygo/lazygo/redis"
	"time"
)

type redisInitiator struct {}

type redisAdapter struct {
	name string
	conn *redis.Redis
}

// init 初始化redis适配器
func (r *redisInitiator) init(opt map[string]string) (Cache, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidRedisAdapterParams
	}

	var err error
	conn, err := redis.Pool(name)
	a := &redisAdapter{
		name:          name,
		conn:          conn,
	}
	return a, err
}

func (r *redisAdapter) Remember(key string, fn func() (interface{}, error), ttl time.Duration) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		err := r.conn.GetObject(key, &wp.Data)
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

		err = r.conn.Set(key, wp.Data, int64(ttl.Seconds()))
		if err != nil {
			return err
		}
		return nil
	}
	return wp
}

func (r *redisAdapter) Set(key string, val interface{}, ttl time.Duration) error {
	wp := &wrapper{}
	err := wp.Pack(val, ttl)
	if err != nil {
		return err
	}
	err = r.conn.Set(key, wp, int64(ttl.Seconds()))
	if err != nil {
		return err
	}
	return nil
}

func (r *redisAdapter) Get(key string) DataResult {
	wp := &wrapper{}
	wp.handler = func(wp *wrapper) error {
		err := r.conn.GetObject(key, &wp.Data)
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

func (r *redisAdapter) Has(key string) (bool, error) {
	wp := &wrapper{}
	err := r.conn.GetObject(key, &wp.Data)
	if err == redigo.ErrNil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return wp.Data.Deadline >= time.Now().Unix(), nil
}

func (r *redisAdapter) Forget(key string) error {
	return r.conn.Del(key)
}

func init() {
	registry.add("redis", &redisInitiator{})
}
