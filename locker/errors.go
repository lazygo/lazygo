package locker

import "errors"

var (
	ErrInvalidRedisAdapterParams = errors.New("invalid redis adapter params")
	ErrInvalidDefaultName        = errors.New("invalid default name")

	ErrAdapterUndefined     = errors.New("undefined adapter")
	ErrAdapterUninitialized = errors.New("uninitialized adapter")
)
