package locker

import (
	"context"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/lazygo/lazygo/redis"
	goredis "github.com/redis/go-redis/v9"
)

// redis适配器
type redisAdapter struct {
	local sync.Map
	name  string
	conn  *goredis.Client
	retry int
}

// UnlockScript 释放锁脚本：仅当持有者匹配时删除，避免误删他人锁
const UnlockScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`

var unlockScript = goredis.NewScript(UnlockScript)

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

// spinInterval 自旋间隔：50ms + 0~50ms随机抖动，避免释放瞬间惊群
func spinInterval() time.Duration {
	return 50*time.Millisecond + time.Duration(rand.IntN(50))*time.Millisecond
}

// Lock 分布式自旋锁
func (r *redisAdapter) Lock(ctx context.Context, resource string, ttl uint64) (Releaser, error) {
	if ttl == 0 {
		return nil, ErrInvalidTTL
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	actual, _ := r.local.LoadOrStore(resource, make(chan struct{}, 1))
	mu := actual.(chan struct{})

	// 进程内串行，排队等待期间响应ctx取消
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case mu <- struct{}{}:
	}

	token := strconv.FormatUint(randomToken(), 10)

	// 获取锁成功，返回释放锁的方法
	// release内部使用带3s超时的独立ctx：acquire阶段的ctx此时可能已过期，不能复用
	var releaseOnce sync.Once
	var releaseErr error
	release := func() error {
		releaseOnce.Do(func() {
			defer func() {
				<-mu
				r.local.Delete(resource)
			}()
			rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			for retry := r.retry; retry > 0; retry-- {
				releaseErr = unlockScript.Run(rctx, r.conn, []string{resource}, token).Err()
				if releaseErr == nil {
					return
				}
			}
		})
		return releaseErr
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	retry := r.retry
	for {
		select {
		case <-ctx.Done():
			_ = release()
			return nil, ctx.Err()
		case <-timer.C:
			timer.Reset(spinInterval())
			// 注意：SET NX未命中时返回false且err为nil，必须取Result判断
			ok, err := r.conn.SetNX(ctx, resource, token, time.Duration(ttl)*time.Second).Result()
			if err != nil {
				if retry > 0 {
					retry--
					continue
				}
				_ = release()
				return nil, err
			}
			if !ok {
				// key已存在，获取锁失败，自旋等待
				runtime.Gosched()
				continue
			}
			return releaseFunc(release), nil
		}
	}
}

// TryLock 获取分布式锁（非阻塞）
func (r *redisAdapter) TryLock(resource string, ttl uint64) (Releaser, bool, error) {
	if ttl == 0 {
		return nil, false, ErrInvalidTTL
	}
	token := strconv.FormatUint(randomToken(), 10)
	// 注意：SET NX未命中时返回false且err为nil，必须取Result判断
	ok, err := r.conn.SetNX(context.Background(), resource, token, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		// key已存在，获取锁失败
		return nil, false, nil
	}

	// 获取锁成功，返回释放锁的方法
	// release内部使用带3s超时的独立ctx，可重复调用（幂等）
	var releaseOnce sync.Once
	var releaseErr error
	release := func() error {
		releaseOnce.Do(func() {
			rctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			for retry := r.retry; retry > 0; retry-- {
				releaseErr = unlockScript.Run(rctx, r.conn, []string{resource}, token).Err()
				if releaseErr == nil {
					return
				}
			}
		})
		return releaseErr
	}

	return releaseFunc(release), true, nil
}

// LockFunc 分布式自旋锁执行fn
func (r *redisAdapter) LockFunc(ctx context.Context, resource string, ttl uint64, f func() any) (result any, err error) {
	lock, err := r.Lock(ctx, resource, ttl)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr := lock.Release(); err == nil {
			err = rerr
		}
	}()
	return f(), nil
}

func init() {
	// 注册适配器
	registry.Add("redis", newRedisLocker)
}
