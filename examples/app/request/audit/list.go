// 查询设备list
package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/sqldb"
)

type ListRequest struct {
	Appid int    `json:"appid" bind:"context"`
	UID   uint64 `json:"uid" bind:"context"`
	sqldb.PaginatorRequest
}

type ListResponse struct {
	sqldb.PaginatorResponse[dbModel.AuditData]
}

func (r *ListRequest) Verify(ctx framework.Context) error {
	r.Clean(sqldb.DefaultPaginatorMaxSize, sqldb.DefaultPaginatorRequest)

	return nil
}
