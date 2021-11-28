package locker

import "sync"

// 分布式锁

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

type Locker interface {

	// Lock 尝试获取锁，会自动自旋一定时间
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	// retry 重试次数 * 重试时间间隔(200ms) 建议大于 超时时间
	Lock(resource string, ttl uint64) (Releaser, error)

	// TryLock 尝试获取锁
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	// retry 重试次数 * 重试时间间隔(200ms) 建议大于 超时时间
	TryLock(resource string, ttl uint64) (Releaser, bool, error)

	// LockFunc 启用分布式锁执行func
	// resource 资源标识，相同的资源标识会互斥
	// ttl 生存时间 (秒)
	// f 返回interface{} 的函数
	// 在获取锁失败或超时的情况下，f不会被执行
	LockFunc(resource string, ttl uint64, fn func() interface{}) (interface{}, error)
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

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}

		a, err := registry.get(item.Adapter)
		if err != nil {
			return err
		}
		lock, err := a.init(item.Option)
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
func Lock(resource string, ttl uint64) (Releaser, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, err
	}
	return lock.Lock(resource, ttl)
}

// TryLock 尝试``1123获取锁
// resource 资源标识，相同的资源标识会互斥
// ttl 生存时间 (秒)
func TryLock(resource string, ttl uint64) (Releaser, bool, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, false, err
	}
	return lock.TryLock(resource, ttl)
}

// LockFunc 启用分布式锁执行func
// resource 资源标识，相同的资源标识会互斥
// ttl 生存时间 (秒)
// f 返回interface{} 的函数
// 在获取锁失败或超时的情况下，f不会被执行
func LockFunc(resource string, ttl uint64, fn func() interface{}) (interface{}, error) {
	lock, err := Instance(manager.defaultName)
	if err != nil {
		return nil, err
	}
	return lock.LockFunc(resource, ttl, fn)
}
