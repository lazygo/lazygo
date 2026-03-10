// 查询设备list
package request

import (
	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/mysql"
)

type ListRequest struct {
	Appid int    `json:"appid" bind:"context"`
	UID   uint64 `json:"uid" bind:"context"`
	mysql.PaginatorRequest
}

type ListResponse struct {
	mysql.PaginatorResponse[dbModel.AuditData]
}

func (r *ListRequest) Verify(ctx framework.Context) error {
	r.Clean(mysql.DefaultPaginatorMaxSize, mysql.DefaultPaginatorRequest)

	return nil
}
