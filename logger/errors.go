package logger

import "errors"

var (
	ErrInvalidDefaultName = errors.New("invalid default name")

	ErrAdapterUndefined     = errors.New("undefined adapter")
	ErrAdapterUninitialized = errors.New("uninitialized adapter")

	ErrInvalidFilename = errors.New("invalid filename")
)
