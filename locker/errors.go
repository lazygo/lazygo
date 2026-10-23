package locker

import "errors"

var (
	ErrInvalidRedisAdapterParams = errors.New("invalid redis adapter params")
	ErrInvalidDefaultName        = errors.New("invalid default name")
	ErrInvalidTTL                = errors.New("ttl must be greater than 0")

	ErrAdapterUninitialized = errors.New("uninitialized adapter")
)
