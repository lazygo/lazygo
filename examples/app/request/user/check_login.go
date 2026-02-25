package request

import (
	"github.com/lazygo/lazygo/examples/framework"
)

type CheckLoginRequest struct {
	UID uint64 `json:"uid" bind:"context"`
}

func (r *CheckLoginRequest) Verify(ctx framework.Context) error {
	return nil
}

func (r *CheckLoginRequest) Clear() {
}
