package locker

import (
	redigo "github.com/gomodule/redigo/redis"
	"github.com/lazygo/lazygo/redis"
	"runtime"
	"strconv"
	"sync"
)

// redis适配器
type redisAdapter struct {
	local sync.Map
	name  string
	conn  *redis.Redis
	retry int
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
		name:  name,
		conn:  conn,
		retry: 3,
	}
	return a, nil
}

// Lock 分布式自旋锁
func (r *redisAdapter) Lock(resource string, ttl uint64) (Releaser, error) {

	actual, _ := r.local.LoadOrStore(resource, &sync.Mutex{})
	mu := actual.(*sync.Mutex)
	mu.Lock()

	token := strconv.FormatUint(randomToken(), 10)

	// 获取锁成功，返回释放锁的方法
	handleRelease := func() error {
		defer func() {
			mu.Unlock()
			r.local.Delete(resource)
		}()
		var err error
		for retry := r.retry; retry >= 0; retry-- {
			_, err = r.conn.Do("EVAL", script, 1, resource, token)
			if err == nil {
				return nil
			}
		}
		return err
	}

	retry := r.retry
	for {
		_, err := r.conn.Do("SET", resource, token, "EX", ttl, "NX")
		if err == redigo.ErrNil {
			// key 已存在，获取锁失败
			runtime.Gosched()
			continue
		}
		if err != nil {
			if retry >= 0 {
				retry--
				continue
			}
			_ = handleRelease()
			mu.Unlock()
			return nil, err
		}
		break
	}

	return releaseFunc(handleRelease), nil
}

// TryLock 获取分布式锁（非阻塞）
func (r *redisAdapter) TryLock(resource string, ttl uint64) (Releaser, bool, error) {
	token := strconv.FormatUint(randomToken(), 10)
	_, err := r.conn.Do("SET", resource, token, "EX", ttl, "NX")
	if err == redigo.ErrNil {
		// key 已存在，获取锁失败
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	// 获取锁成功，返回释放锁的方法
	handleRelease := func() error {
		var err error
		for retry := r.retry; retry > 0; retry-- {
			_, err = r.conn.Do("EVAL", script, 1, resource, token)
			if err == nil {
				return nil
			}
		}
		return err
	}

	return releaseFunc(handleRelease), true, nil
}

func (r *redisAdapter) LockFunc(resource string, ttl uint64, fn func() interface{}) (result interface{}, err error) {
	lock, err := r.Lock(resource, ttl)
	if err != nil {
		return nil, err
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
