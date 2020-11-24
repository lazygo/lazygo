package locker

import (
	"errors"
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

// 初始化redis适配器
func (r *redisAdapter) init(config map[string]interface{}) error {
	if _, ok := config["name"]; !ok {
		return errors.New("redis适配器参数错误")
	}
	r.name = toString(config["name"])
	p, err := redis.RedisPool(r.name)
	if err != nil {
		return err
	}
	r.conn = p
	if retryInterval, ok := config["retry_interval"]; ok {
		r.retryInterval = uint64(toInt64(retryInterval))
	}
	return nil
}

// lock 获取锁
func (r *redisAdapter) lock(resource string, ttl time.Duration, retry uint) (Locker, bool, error) {

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

	// 获取锁成功
	handleRelease := func() error {
		// 返回释放锁的方法
		_, err := r.conn.Do("EVAL", script, 1, resource, token)
		return err
	}
	return releaseFunc(handleRelease), true, nil
}

func init() {
	// 注册适配器
	registry["redis"] = &redisAdapter{
		retryInterval: 200, // 默认200毫秒重试间隔（毫秒）
	}
}
