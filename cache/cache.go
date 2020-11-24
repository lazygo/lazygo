package cache

import (
	"errors"
	"time"
)

type DataResult interface {
	GetType() int32
	ToString() (string, error)
	ToByteArray() ([]byte, error)
	ToMap() (Map, error)
	ToMapSlice() (MapSlice, error)
}

type Cache interface {
	Remember(key string, value interface{}, ttl time.Duration) DataResult
	Get(key string) (DataResult, error)
	Has(key string) bool
	Forget(key string) bool
}

type adapter interface {
	Cache
	init(map[string]interface{}) error
}

// 适配器注册器
var registry = map[string]adapter{}

// 初始化配置
var setting *typeSetting = nil

type typeSetting struct {
	adapterName    string
}

// Init 初始化设置
// 在框架初始化时调用
// adapterName 适配器名称
// ext 选项
func Init(adapterName string, opt map[string]interface{}) {
	if setting != nil {
		panic(errors.New("不能重复初始化缓存"))
	}

	setting = &typeSetting{
		adapterName:    adapterName,
	}

	if adapter, ok := registry[adapterName]; ok {
		err := adapter.init(opt)
		if err != nil {
			panic(err)
		}
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// Remember 设置缓存
func Remember(key string, value interface{}, ttl time.Duration) DataResult {
	if setting == nil {
		panic(errors.New("缓存未初始化"))
	}
	// 获取适配器
	adapterName := setting.adapterName
	if adapter, ok := registry[adapterName]; ok {
		return adapter.Remember(key, value, ttl)
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// Get 获取缓存
func Get(key string) (DataResult, error) {
	if setting == nil {
		panic(errors.New("缓存未初始化"))
	}
	// 获取适配器
	adapterName := setting.adapterName
	if adapter, ok := registry[adapterName]; ok {
		return adapter.Get(key)
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// Has 判断缓存是否存在
func Has(key string) bool {
	if setting == nil {
		panic(errors.New("缓存未初始化"))
	}
	// 获取适配器
	adapterName := setting.adapterName
	if adapter, ok := registry[adapterName]; ok {
		return adapter.Has(key)
	}
	panic(errors.New("找不到适配器" + adapterName))
}

// Forget 删除缓存
func Forget(key string) bool {
	if setting == nil {
		panic(errors.New("缓存未初始化"))
	}
	// 获取适配器
	adapterName := setting.adapterName
	if adapter, ok := registry[adapterName]; ok {
		return adapter.Forget(key)
	}
	panic(errors.New("找不到适配器" + adapterName))
}