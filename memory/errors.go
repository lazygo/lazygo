package memory

import "errors"

var (
	ErrLRUCacheNotExists = errors.New("指定LRUCache不存在，或未初始化")
	ErrSizeOverflow      = errors.New("LRU容量溢出")
)
