package memcache

import "errors"

var (
	ErrMemcacheNotExists = errors.New("指定Memcache不存在，或未初始化")
)
