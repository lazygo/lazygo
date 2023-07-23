package locker

import (
	"context"
	"runtime"
	"strconv"
	"sync"
	"time"
	"unsafe"

	goredis "github.com/go-redis/redis/v8"
	"github.com/lazygo/lazygo/redis"
)

// redis适配器
type redisAdapter struct {
	local sync.Map
	name  string
	conn  *goredis.Client
	retry int
}

// UNLOCK_SCRIPT 释放锁脚本
const UNLOCK_SCRIPT = `
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
	conn, err := redis.Client(name)
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
func (r *redisAdapter) Lock(ctx context.Context, resource string, ttl uint64) (Releaser, error) {

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
			err = goredis.NewScript(UNLOCK_SCRIPT).Run(context.Background(), r.conn, []string{resource}, token).Err()
			if err == nil {
				return nil
			}
		}
		return err
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	retry := r.retry
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			timer.Reset(50 * time.Millisecond)
			err := r.conn.SetNX(ctx, resource, token, time.Duration(ttl)*time.Second).Err()
			if err != nil {
				if err == goredis.Nil {
					// key 已存在，获取锁失败
					runtime.Gosched()
					continue
				}
				if retry >= 0 {
					retry--
					continue
				}
				_ = handleRelease()
				return nil, err
			}
			return releaseFunc(handleRelease), nil
		}
	}
}

// TryLock 获取分布式锁（非阻塞）
func (r *redisAdapter) TryLock(resource string, ttl uint64) (Releaser, bool, error) {
	token := strconv.FormatUint(randomToken(), 10)
	err := r.conn.SetNX(context.Background(), resource, token, time.Duration(ttl)*time.Second).Err()
	if err == goredis.Nil {
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
			err = goredis.NewScript(UNLOCK_SCRIPT).Run(context.Background(), r.conn, []string{resource}, token).Err()
			if err == nil {
				return nil
			}
		}
		return err
	}

	return releaseFunc(handleRelease), true, nil
}

func (r *redisAdapter) LockFunc(ctx context.Context, ttl uint64, f func() interface{}) (result interface{}, err error) {
	resource := runtime.FuncForPC(**(**uintptr)(unsafe.Pointer(&f))).Name()
	lock, err := r.Lock(ctx, resource, ttl)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = lock.Release()
	}()
	return f(), nil
}

func init() {
	// 注册适配器
	registry.add("redis", adapterFunc(newRedisLocker))
}
