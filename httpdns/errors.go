package httpdns

import "errors"

var (
	ErrInvalidHTTPDNSAdapterParams = errors.New("invalid httpdns adapter params")
	ErrInvalidDefaultName          = errors.New("invalid default name")
	ErrAdapterUninitialized        = errors.New("uninitialized adapter")
)
