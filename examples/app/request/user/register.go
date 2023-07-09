package request

import (
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

type RegisterRequest struct {
	Username string `json:"username" bind:"form"`
	Password string `json:"password" bind:"form"`
}

func (r *RegisterRequest) Verify() error {
	if !utils.EmailRegex.MatchString(r.Username) && !utils.MobileRegex.MatchString(r.Username) {
		return errors.ErrUsernameError
	}
	if len(r.Password) < 6 {
		return errors.ErrPasswordError
	}
	if len(r.Password) > 32 {
		return errors.ErrPasswordTooLongError
	}
	return nil
}

func (r *RegisterRequest) Clear() {
}
