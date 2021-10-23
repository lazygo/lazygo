package cache

import "errors"

var (
	ErrInvalidRedisAdapterParams = errors.New("无效的redis适配器参数")
	ErrInvalidMemcacheAdapterParams = errors.New("无效的memcache适配器参数")

	ErrAdapterNotFound = errors.New("找不到适配器")
	ErrAdapterUninitialized = errors.New("适配器未初始化")

	ErrEmptyKey = errors.New("key不存在或已过期")
)
