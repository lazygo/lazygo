package locker

// redis适配器测试，需要本地redis：
//   REDIS_ADDR     默认 127.0.0.1:6379
//   REDIS_PASSWORD 默认空

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lazygo/lazygo/redis"
	testify "github.com/stretchr/testify/assert"
)

func setupLocker(t *testing.T) {
	assert := testify.New(t)

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	host, port, found := strings.Cut(addr, ":")
	if !found {
		host, port = addr, "6379"
	}
	var portNum int
	fmt.Sscanf(port, "%d", &portNum)

	err := redis.Init([]redis.Config{
		{
			Name:     "locker_unit_test",
			Host:     host,
			Port:     portNum,
			Password: os.Getenv("REDIS_PASSWORD"),
			Db:       0,
		},
	})
	assert.Nil(err)

	// 验证连通性，redis不可用（未启动或密码错误）时跳过
	rdb, err := redis.Client("locker_unit_test")
	if err != nil {
		t.Skipf("redis %s 不可用: %v", addr, err)
	}
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis %s 不可用: %v", addr, err)
	}

	err = Init([]Config{
		{
			Name:    "locker_unit_test",
			Adapter: "redis",
			Option:  map[string]string{"name": "locker_unit_test"},
		},
	}, "locker_unit_test")
	assert.Nil(err)
}

// 同资源TryLock第二次必须失败，释放后可重新获取
func TestTryLockContention(t *testing.T) {
	assert := testify.New(t)
	setupLocker(t)
	res := t.Name()

	r1, ok1, err := TryLock(res, 30)
	assert.Nil(err)
	assert.True(ok1)

	r2, ok2, err := TryLock(res, 30)
	assert.Nil(err)
	assert.False(ok2)
	assert.Nil(r2)

	// 释放第一个后，资源应可再次获取
	assert.Nil(r1.Release())
	r3, ok3, err := TryLock(res, 30)
	assert.Nil(err)
	assert.True(ok3)
	assert.Nil(r3.Release())
}

// 多协程竞争同一把锁，临界区不允许重叠
func TestLockSerialization(t *testing.T) {
	assert := testify.New(t)
	setupLocker(t)
	res := t.Name()

	var occupied int32
	var violations int32
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rel, err := Lock(context.Background(), res, 30)
			if err != nil {
				t.Errorf("Lock error: %v", err)
				return
			}
			defer rel.Release()

			if !atomic.CompareAndSwapInt32(&occupied, 0, 1) {
				atomic.AddInt32(&violations, 1)
			}
			time.Sleep(50 * time.Millisecond)
			atomic.StoreInt32(&occupied, 0)
		}()
	}
	wg.Wait()
	assert.Equal(int32(0), atomic.LoadInt32(&violations))
}

// 锁被持有时，带超时的Lock应返回ctx错误且未持有锁
func TestLockTimeout(t *testing.T) {
	assert := testify.New(t)
	setupLocker(t)
	res := t.Name()

	rel, err := Lock(context.Background(), res, 30)
	assert.Nil(err)
	defer rel.Release()

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	rel2, err := Lock(ctx, res, 30)
	assert.Nil(rel2)
	assert.ErrorIs(err, context.DeadlineExceeded)
	assert.GreaterOrEqual(time.Since(start), 300*time.Millisecond)

	// 超时方未持有锁：资源仍被第一个持有者占用
	_, ok2, err := TryLock(res, 30)
	assert.Nil(err)
	assert.False(ok2)
}

func TestLockFunc(t *testing.T) {
	assert := testify.New(t)
	setupLocker(t)
	res := t.Name()

	// 正常执行并返回结果
	result, err := LockFunc(context.Background(), res, 30, func() (string, error) {
		return "hello", nil
	})
	assert.Nil(err)
	assert.Equal("hello", result)

	// fn返回错误时透传，且锁已释放
	sentinel := errors.New("fn boom")
	_, err = LockFunc(context.Background(), res, 30, func() (string, error) {
		return "", sentinel
	})
	assert.ErrorIs(err, sentinel)

	r, ok, err := TryLock(res, 30)
	assert.Nil(err)
	assert.True(ok)
	assert.Nil(r.Release())
}

// ttl必须大于0，否则创建出永不过期的锁（崩溃即死锁）
func TestInvalidTTL(t *testing.T) {
	assert := testify.New(t)
	setupLocker(t)
	res := t.Name()

	_, ok, err := TryLock(res, 0)
	assert.Equal(ErrInvalidTTL, err)
	assert.False(ok)

	_, err = Lock(context.Background(), res, 0)
	assert.Equal(ErrInvalidTTL, err)

	_, err = LockFunc(context.Background(), res, 0, func() (string, error) { return "", nil })
	assert.Equal(ErrInvalidTTL, err)
}
