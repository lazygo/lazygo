// 查询设备list
package request

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/model"
	dbModel "github.com/lazygo-dev/lazygo/examples/model/db"
)

type ListRequest struct {
	Appid int    `json:"appid" bind:"context"`
	UID   uint64 `json:"uid" bind:"context"`
	model.PaginatorRequest
}

type ListResponse struct {
	model.PaginatorResponse[dbModel.AuditData]
}

func (r *ListRequest) Verify(ctx framework.Context) error {
	r.Clean(model.DefaultPaginatorMaxSize, model.DefaultPaginatorRequest)

	return nil
}

func (r *ListRequest) Clear() {
}
