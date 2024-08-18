package locker

import "errors"

var (
	ErrInvalidRedisAdapterParams = errors.New("invalid redis adapter params")
	ErrInvalidDefaultName        = errors.New("invalid default name")

	ErrAdapterUninitialized = errors.New("uninitialized adapter")
)
