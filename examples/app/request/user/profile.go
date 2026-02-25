package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
)

type ProfileRequest struct {
	UID uint64 `json:"uid" bind:"context"`
}

type ProfileResponse struct {
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	Mobile       string `json:"mobile"`
	Email        string `json:"email"`
	RegisterTime string `json:"register_time"`
}

func (r *ProfileRequest) Verify(ctx framework.Context) error {

	return nil
}

func (r *ProfileRequest) Clear() {
}
