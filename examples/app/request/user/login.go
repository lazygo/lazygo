package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

type LoginRequest struct {
	Username string `json:"username" bind:"query,form" process:"trim"`
	Password string `json:"password" bind:"query,form" process:"trim,cut(32)"`
	Type     int
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (r *LoginRequest) Verify(ctx framework.Context) error {
	r.Type = utils.UsernameType(r.Username)
	if r.Type != utils.TypeEmail && r.Type != utils.TypeMobile {
		return errors.ErrUsernameInvalid
	}

	return nil
}

func (r *LoginRequest) Clear() {
}
