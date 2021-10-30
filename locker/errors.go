package locker

import "errors"

var (
	ErrInvalidRedisAdapterParams = errors.New("无效的redis适配器参数")
	ErrInvalidDefaultName        = errors.New("无效的默认实例")

	ErrAdapterNotFound      = errors.New("找不到适配器")
	ErrAdapterUninitialized = errors.New("适配器未初始化")

	ErrTimeout = errors.New("获取锁超时")
)
