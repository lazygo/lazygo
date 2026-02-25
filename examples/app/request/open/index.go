package debug

import (
	"github.com/lazygo/lazygo/examples/framework"
)

type IndexRequest struct {
	Appid int `json:"appid" bind:"context"`
}

func (r *IndexRequest) Verify(ctx framework.Context) error {

	return nil
}

func (r *IndexRequest) Clear() {
}
