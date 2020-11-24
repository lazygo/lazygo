package locker

import (
	"errors"
	"time"
)

// 分布式锁

// 适配器接口
type adapter interface {
	init(map[string]interface{}) error
	lock(resource string, ttl time.Duration, retry uint) (Locker, bool, error)
}

// 释放锁接口
type Locker interface {
	Release() error
}

// 释放锁包装器
type releaseFunc func() error

func (r releaseFunc) Release() error { return r() }

// 适配器注册器
var registry = map[string]adapter{}

// 初始化配置
var setting *typeSetting = nil

type typeSetting struct {
	adapterName    string
	defaultTimeout time.Duration
}

// 初始化设置
func Init(adapterName string, timeout time.Duration, ext map[string]interface{}) {
	if setting != nil {
		panic(errors.New("不能重复初始化分布式锁"))
	}

	setting = &typeSetting{
		adapterName:    adapterName,
		defaultTimeout: timeout,
	}

	if adapter, ok := registry[adapterName]; ok {
		err := adapter.init(ext)
		if err != nil {
			panic(err)
		}
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// TryLock 尝试获取锁
// resource 资源标识，相同的资源标识会互斥
// ttl 生存时间 (秒)
// retry 重试次数 * 重试时间间隔(200ms) 建议大于 超时时间
func TryLock(resource string, ttl int, retry uint) (Locker, bool, error) {
	if setting == nil {
		panic(errors.New("分布式锁未初始化"))
	}
	// 获取适配器
	adapterName := setting.adapterName
	if adapter, ok := registry[adapterName]; ok {
		return adapter.lock(resource, time.Duration(ttl) * time.Second, retry)
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// LockFunc 启用分布式锁执行func
// resource 资源标识，相同的资源标识会互斥
// ttl 生存时间 (秒)
// f 返回interface{} 的函数
// 在获取锁失败或超时的情况下，f不会被执行
func LockFunc(resource string, ttl int, f func() interface{}) (interface{}, error) {
	// 重试次数 * 重试时间间隔 应大于 超时时间
	retry := uint(ttl * 10)
	lock, ok, err := TryLock(resource, ttl, retry)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("获取锁超时")
	}
	defer lock.Release()
	return f(), nil
}
