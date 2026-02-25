package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/utils"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
)

type BindMobileRequest struct {
	UID     uint64 `json:"uid" bind:"ctx"`
	Mobile  string `json:"mobile" bind:"query,form" process:"trim,tolower"`
	Captcha string `json:"captcha" bind:"query,form" process:"trim,tolower,cut(6)"`
}

type BindMobileResponse struct {
}

func (r *BindMobileRequest) Verify(ctx framework.Context) error {
	if utils.UsernameType(r.Mobile) != utils.TypeMobile {
		return errors.ErrInvalidMobile
	}
	return nil
}

func (r *BindMobileRequest) Clear() {
}
