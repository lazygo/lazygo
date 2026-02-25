package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	cacheModel "github.com/lazygo-dev/lazygo/examples/model/cache"
	"github.com/lazygo-dev/lazygo/examples/utils"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
)

type CaptchaRequest struct {
	UID      uint64 `json:"uid" bind:"ctx"`
	Username string `json:"username" bind:"query,form" process:"trim,tolower"`
	Opname   string `json:"opname" bind:"query,form" process:"trim,tolower,cut(16)"`
	Type     int
}

type CaptchaResponse struct {
}

func (r *CaptchaRequest) Verify(ctx framework.Context) error {
	r.Type = utils.UsernameType(r.Username)
	if r.Type != utils.TypeEmail && r.Type != utils.TypeMobile {
		return errors.ErrUsernameInvalid
	}
	if _, ok := cacheModel.SmsTpl[r.Opname]; !ok {
		return errors.ErrInvalidParams
	}

	return nil
}

func (r *CaptchaRequest) Clear() {
}
