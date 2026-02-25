package request

import "github.com/lazygo/lazygo/examples/framework"

type ConnectionRequest struct {
	UID uint64 `json:"uid" bind:"context"`
}

func (r *ConnectionRequest) Verify(ctx framework.Context) error {

	return nil
}

func (r *ConnectionRequest) Clear() {
}
