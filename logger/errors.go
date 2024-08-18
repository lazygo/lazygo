package logger

import "errors"

var (
	ErrInvalidDefaultName   = errors.New("invalid default name")
	ErrAdapterUninitialized = errors.New("uninitialized adapter")
	ErrInvalidFilename      = errors.New("invalid filename")
)
