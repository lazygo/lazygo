package controller

import (
	"encoding/json"
	"fmt"
	"slices"

	request "github.com/lazygo/lazygo/examples/app/request/audit"
	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/pkg/goutils"
)

// AuditController 审计
type AuditController struct {
	Ctx framework.Context
}

func (ctl *AuditController) List(req *request.ListRequest) (*request.ListResponse, error) {
	mdlAudit := dbModel.NewAuditModel(ctl.Ctx)
	cond := map[string]any{
		"uid": req.UID,
	}
	data, _, err := mdlAudit.Paginator(cond, &req.PaginatorRequest, nil)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: list dev failed] [err: %v]", err)
		return nil, err
	}

	for index, item := range data.List {
		var contentMap map[string]any
		err = json.Unmarshal([]byte(item.Content), &contentMap)
		if err != nil {
			ctl.Ctx.Logger().Warn("[msg: unmarshal audit log content failed] [content: %v] [err: %v]", item.Content, err)
			continue
		}
		for k, v := range contentMap {
			if slices.Contains([]string{"uid", "id", "relate_id", "ip", "mtime"}, k) {
				delete(contentMap, k)
			}
			if slices.Contains([]string{"username", "mobile", "email", "phone", "ip"}, k) {
				contentMap[k] = goutils.Mask(fmt.Sprintf("%v", v), 3, 3)
			}
		}
		contentJson, err := json.Marshal(contentMap)
		if err != nil {
			ctl.Ctx.Logger().Warn("[msg: marshal audit log content failed] [err: %v]", err)
			continue
		}
		data.List[index].Content = string(contentJson)
	}

	resp := &request.ListResponse{PaginatorResponse: *data}

	return resp, nil
}
