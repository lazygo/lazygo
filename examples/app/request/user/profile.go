package request

import "github.com/lazygo/lazygo/examples/framework"

type ProfileRequest struct {
	UID uint64 `json:"uid" bind:"ctx"`
}

type ProfileResponse struct {
}

func (r *ProfileRequest) Verify(ctx framework.Context) error {
	return nil
}

func (r *ProfileRequest) Clear() {
}
