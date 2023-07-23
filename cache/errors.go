package cache

import "errors"

var (
	ErrInvalidRedisAdapterParams    = errors.New("invalid redis adapter params")
	ErrInvalidMemcacheAdapterParams = errors.New("invalid memcache adapter params")
	ErrInvalidLruAdapterParams      = errors.New("invalid lru adapter params")
	ErrInvalidDefaultName           = errors.New("invalid default name")

	ErrAdapterUndefined     = errors.New("undefined adapter")
	ErrAdapterUninitialized = errors.New("uninitialized adapter")
)
