package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lazygo/lazygo/utils"
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrTooManyRequests             = NewHTTPError(http.StatusTooManyRequests)
	ErrBadRequest                  = NewHTTPError(http.StatusBadRequest)
	ErrBadGateway                  = NewHTTPError(http.StatusBadGateway)
	ErrInternalServerError         = NewHTTPError(http.StatusInternalServerError)
	ErrRequestTimeout              = NewHTTPError(http.StatusRequestTimeout)
	ErrServiceUnavailable          = NewHTTPError(http.StatusServiceUnavailable)
	ErrValidatorNotRegistered      = errors.New("validator not registered")
	ErrActionNotExists             = errors.New("action not exists")
	ErrInvalidRedirectCode         = errors.New("invalid redirect status code")
	ErrCookieNotFound              = errors.New("cookie not found")
	ErrInvalidListenerNetwork      = errors.New("invalid listener network")
)

// Error handlers
var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
	}
)

// HTTPError represents an error that occurred while handling a request.
type HTTPError struct {
	Code     int   `json:"code"`
	Errno    int   `json:"errno"`
	Message  any   `json:"message"`
	Internal error `json:"-"` // Stores the error returned by an external dependency
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...any) *HTTPError {
	he := &HTTPError{Code: code, Errno: code, Message: http.StatusText(code)}
	switch len(message) {
	case 0:
	case 1:
		he.Message = message[0]
	case 2:
		he.Errno = utils.ToInt(message[0], -1)
		he.Message = message[1]
	default:
		he.Message = message[0]
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	if he.Internal == nil {
		return fmt.Sprintf("code=%d, errno=%d, message=%v", he.Code, he.Errno, he.Message)
	}
	return fmt.Sprintf("code=%d, errno=%d, message=%v, internal=%v", he.Code, he.Errno, he.Message, he.Internal)
}

// SetInternal copy and set error to HTTPError.Internal
func (he HTTPError) SetInternal(err error) *HTTPError {
	he.Internal = err
	return &he
}

// Unwrap satisfies the Go 1.13 error wrapper interface.
func (he *HTTPError) Unwrap() error {
	return he.Internal
}
