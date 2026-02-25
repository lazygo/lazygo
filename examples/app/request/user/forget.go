package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/utils"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
)

type ForgetRequest struct {
	Username string `json:"username" bind:"query,form" process:"trim,tolower"`
	Captcha  string `json:"captcha" bind:"query,form" process:"trim,tolower,cut(6)"`
	Password string `json:"password" bind:"query,form" process:"trim"`
	Type     int
}

func (r *ForgetRequest) Verify(ctx framework.Context) error {
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

func (r *ForgetRequest) Clear() {
}
