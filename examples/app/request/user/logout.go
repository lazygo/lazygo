package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
)

type LogoutRequest struct {
	Authorization string `json:"authorization" bind:"header" process:"trim,cut(32)"`
}

type LogoutResponse struct {
}

func (r *LogoutRequest) Verify(ctx framework.Context) error {
	return nil
}

func (r *LogoutRequest) Clear() {
}
