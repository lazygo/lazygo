package request

import (
	"cmp"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

// 设备登录时复用此方法，即支持json请求也支持来自设备的formdata请求
type LoginRequest struct {
	Referer   string `json:"referer" bind:"cookie" process:"trim"`
	Origin    string `json:"origin" bind:"cookie" process:"trim"`
	Username  string `json:"username" bind:"query,form" process:"trim,tolower"`
	Password  string `json:"password" bind:"query,form" process:"trim,cut(32)"`
	Loginpass string `json:"loginpass"`
	Type      int
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (r *LoginRequest) Verify(ctx framework.Context) error {
	r.Type = utils.UsernameType(r.Username)
	if r.Type != utils.TypeEmail && r.Type != utils.TypeMobile {
		return errors.ErrUsernameInvalid
	}

	r.Password = cmp.Or(r.Password, r.Loginpass)
	return nil
}

func (r *LoginRequest) Clear() {
}
