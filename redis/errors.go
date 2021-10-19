package redis

import "errors"

var (
	ErrRedisNotExists = errors.New("指定Redis不存在，或未初始化")
)
