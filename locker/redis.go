package locker

import (
	redigo "github.com/gomodule/redigo/redis"
	"github.com/lazygo/lazygo/redis"
	"strconv"
	"time"
)

// redis适配器
type redisAdapter struct {
	name          string
	conn          *redis.Redis
	retryInterval uint64
}

// 释放锁脚本
const script = `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `

// newRedisLocker 初始化redis适配器
func newRedisLocker(opt map[string]string) (Locker, error) {
	name, ok := opt["name"]
	if !ok || name == "" {
		return nil, ErrInvalidRedisAdapterParams
	}

	var err error
	conn, err := redis.Pool(name)
	if err != nil {
		return nil, err
	}
	a := &redisAdapter{
		name:          name,
		conn:          conn,
		retryInterval: 100, // 默认100毫秒重试间隔（毫秒）
	}
	if retryInterval, ok := opt["retry_interval"]; ok {
		a.retryInterval, err = strconv.ParseUint(retryInterval, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

// lock 获取锁
func (r *redisAdapter) lock(resource string, ttl time.Duration, retry uint64) (Releaser, bool, error) {

	token := strconv.FormatUint(randomToken(), 10)
	for ; retry >= 0; retry-- {
		_, err := r.conn.Do("SET", resource, token, "EX", ttl.Seconds(), "NX")
		if err == redigo.ErrNil {
			// key 已存在，获取锁失败
			if retry > 0 && r.retryInterval > 0 {
				// 等待
				delay := time.Duration(randRange(r.retryInterval/2, r.retryInterval)) * time.Millisecond
				<-time.After(delay)
				continue
			}
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
	}

	// 获取锁成功，返回释放锁的方法
	handleRelease := func() error {
		var err error
		for retry = 5; retry > 0; retry-- {
			_, err = r.conn.Do("EVAL", script, 1, resource, token)
			if err == nil {
				return nil
			}
		}
		return err
	}

	return releaseFunc(handleRelease), true, nil
}

func (r *redisAdapter) Lock(resource string, ttl uint64) (Releaser, bool, error) {
	return r.lock(resource, time.Duration(ttl)*time.Second, 0)
}

func (r *redisAdapter) TryLock(resource string, ttl uint64, retry uint64) (Releaser, bool, error) {
	return r.lock(resource, time.Duration(ttl)*time.Second, retry)
}

func (r *redisAdapter) LockFunc(resource string, ttl uint64, fn func() interface{}) (result interface{}, err error) {
	// 重试次数 * 重试时间间隔 应大于 超时时间
	retry := 10 * ttl * 1000 / r.retryInterval
	lock, ok, err := r.TryLock(resource, ttl, retry)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrTimeout
	}
	defer func() {
		err = lock.Release()
	}()
	return fn(), nil
}

func init() {
	// 注册适配器
	registry.add("redis", adapterFunc(newRedisLocker))
}
