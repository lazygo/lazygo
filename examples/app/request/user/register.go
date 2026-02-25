package request

import (
	"cmp"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

type RegisterRequest struct {
	Referer   string `json:"referer" bind:"cookie" process:"trim"`
	Origin    string `json:"origin" bind:"cookie" process:"trim"`
	Username  string `json:"username" bind:"query,form" process:"trim,tolower"`
	Captcha   string `json:"captcha" bind:"query,form" process:"trim,tolower,cut(6)"`
	Password  string `json:"password" bind:"query,form" process:"trim"`
	Loginpass string `json:"loginpass"`
	Type      int
}

func (r *RegisterRequest) Verify(ctx framework.Context) error {
	r.Type = utils.UsernameType(r.Username)
	if r.Type != utils.TypeEmail && r.Type != utils.TypeMobile {
		return errors.ErrUsernameInvalid
	}
	r.Password = cmp.Or(r.Password, r.Loginpass)
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
