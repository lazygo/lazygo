package logger

import "errors"

var (
	ErrInvalidDefaultName = errors.New("无效的默认实例")

	ErrAdapterNotFound      = errors.New("找不到适配器")
	ErrAdapterUninitialized = errors.New("适配器未初始化")
	ErrInvalidFilename      = errors.New("文件名错误")
)
