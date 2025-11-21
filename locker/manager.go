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
	// ttl 生存时间 (秒)
	Lock(ctx context.Context, resource string, ttl uint64) (Releaser, error)

	// TryLock 尝试获取锁
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	TryLock(resource string, ttl uint64) (Releaser, bool, error)

	// LockFunc 分布式自旋锁执行fn
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	// f 返回interface{} 的函数
	// 在获取锁失败或超时的情况下，fn不会被执行
	LockFunc(ctx context.Context, ttl uint64, fn func() any) (any, error)
}

type Releaser interface {
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
func Init(conf []Config, defaultAdapter string) error {
	return manager.init(conf, defaultAdapter)
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

// LockFunc 启用分布式锁执行fn
// ttl 生存时间 (秒)
// f 返回interface{} 的函数
// 在获取锁失败或超时的情况下，fn不会被执行
func LockFunc(ctx context.Context, ttl uint64, fn func() any) (any, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, err
	}
	return lock.LockFunc(ctx, ttl, fn)
}
