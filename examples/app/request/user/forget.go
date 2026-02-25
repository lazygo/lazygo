package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/pkg/goutils"
)

type ForgetRequest struct {
	Username string `json:"username" bind:"query,form" process:"trim,tolower"`
	Captcha  string `json:"captcha" bind:"query,form" process:"trim,tolower,cut(6)"`
	Password string `json:"password" bind:"query,form" process:"trim"`
	Type     int
}

func (r *ForgetRequest) Verify(ctx framework.Context) error {
	r.Type = goutils.UsernameType(r.Username)
	if r.Type != goutils.TypeEmail && r.Type != goutils.TypeMobile {
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
