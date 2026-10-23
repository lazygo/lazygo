package locker

import (
	"context"
	"sync"

	"github.com/lazygo/lazygo/internal"
)

// 分布式锁

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

type Locker interface {

	// Lock 分布式自旋锁
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)，必须大于0
	Lock(ctx context.Context, resource string, ttl uint64) (Releaser, error)

	// TryLock 尝试获取锁
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	TryLock(resource string, ttl uint64) (Releaser, bool, error)

	// LockFunc 分布式自旋锁执行fn
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)，必须大于0
	// 在获取锁失败或超时的情况下，fn不会被执行
	LockFunc(ctx context.Context, resource string, ttl uint64, fn func() any) (any, error)
}

type Releaser interface {
	// Release 释放锁
	// 内部使用带3s超时的独立ctx执行，可重复调用（幂等）
	Release() error
}

type releaseFunc func() error

func (r releaseFunc) Release() error { return r() }

type Manager struct {
	sync.Map
	defaultName string
}

var registry = internal.Register[Locker, map[string]string]{}

var manager = &Manager{}

// init 初始化分布式锁
func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}

		a, err := registry.Get(item.Adapter)
		if err != nil {
			return err
		}
		lock, err := a.Init(item.Option)
		if err != nil {
			return err
		}
		m.Store(item.Name, lock)

		if defaultName == item.Name {
			m.defaultName = defaultName
		}
	}
	if m.defaultName == "" {
		return ErrInvalidDefaultName
	}
	return nil
}

// Init 初始化设置，在框架初始化时调用
func Init(conf []Config, defaultName string) error {
	return manager.init(conf, defaultName)
}

// Instance 获取分布式锁实例
func Instance(name string) (Locker, error) {
	a, ok := manager.Load(name)
	if !ok {
		return nil, ErrAdapterUninitialized
	}
	return a.(Locker), nil
}

// Lock 分布式自旋锁
func Lock(ctx context.Context, resource string, ttl uint64) (Releaser, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, err
	}
	return lock.Lock(ctx, resource, ttl)
}

// TryLock 尝试获取锁
// resource 资源标识，相同的资源标识会互斥
// ttl 生存时间 (秒)
func TryLock(resource string, ttl uint64) (Releaser, bool, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, false, err
	}
	return lock.TryLock(resource, ttl)
}

// LockFunc 启用分布式锁执行fn，fn支持返回错误
// 获取锁失败或超时的情况下，fn不会被执行
// fn返回错误时锁照常释放；释放失败且fn未报错时返回释放错误
func LockFunc[T any](ctx context.Context, resource string, ttl uint64, fn func() (T, error)) (result T, err error) {
	lock, err := Lock(ctx, resource, ttl)
	if err != nil {
		return result, err
	}
	defer func() {
		if rerr := lock.Release(); err == nil {
			err = rerr
		}
	}()
	return fn()
}

// CloseAll 清空已注册的锁实例
// 注意底层redis连接由redis包管理，不在此关闭
func CloseAll() {
	manager.Range(func(name, _ any) bool {
		manager.Delete(name)
		return true
	})
	manager.defaultName = ""
}
