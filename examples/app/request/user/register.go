package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

type RegisterRequest struct {
	Username string `json:"username" bind:"query,form" process:"trim"`
	Password string `json:"password" bind:"query,form" process:"trim,cut(32)"`
	Type     int
}

func (r *RegisterRequest) Verify(ctx framework.Context) error {
	r.Type = utils.UsernameType(r.Username)
	if r.Type != utils.TypeEmail && r.Type != utils.TypeMobile {
		return errors.ErrUsernameInvalid
	}
	if len(r.Password) < 6 {
		return errors.ErrPasswordTooShort
	}
	if len(r.Password) > 32 {
		return errors.ErrPasswordTooLong
	}
	return nil
}

func (r *RegisterRequest) Clear() {
}
