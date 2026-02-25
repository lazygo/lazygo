package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/pkg/goutils"
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
	r.Type = goutils.UsernameType(r.Username)
	if r.Type != goutils.TypeEmail && r.Type != goutils.TypeMobile {
		return errors.ErrUsernameInvalid
	}
	if _, ok := cacheModel.SmsTpl[r.Opname]; !ok {
		return errors.ErrInvalidParams
	}

	return nil
}

func (r *CaptchaRequest) Clear() {
}
