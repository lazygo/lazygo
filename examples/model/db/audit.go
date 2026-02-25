package db

import (
	"encoding/json"

	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/model"
)

const (
	AuditTypeLogin          = "login"
	AuditTypeRegister       = "register"
	AuditTypeForgetPassword = "forget"
	AuditTypeBindMobile     = "bind_mobile"
	AuditTypeSendCaptcha    = "send_captcha"
)

// AuditModel 审计日志
type AuditModel struct {
	Ctx framework.Context
	model.TxModel[AuditData]
}

type AuditData struct {
	UID      uint64 `json:"uid"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	RelateID string `json:"relate_id"`
	IP       string `json:"ip"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

func NewAuditModel(ctx framework.Context) *AuditModel {
	mdl := &AuditModel{Ctx: ctx}
	mdl.SetTable("audit")
	mdl.SetDB("lazygo-db")
	return mdl
}

func (mdl *AuditModel) Log(uid uint64, opType string, data any, relateID string) {
	content, _ := json.Marshal(data)
	_, err := mdl.Insert(map[string]any{
		"uid":       uid,
		"type":      opType,
		"content":   string(content),
		"relate_id": relateID,
		"ip":        mdl.Ctx.RealIP(),
	})
	if err != nil {
		mdl.Ctx.Logger().Warn("[msg: add user audit log fail] [err: %v]", err)
	}
}

func (mdl *AuditModel) Get(uid uint64, opType string) (*AuditData, int, error) {
	cond := map[string]any{
		"uid":  uid,
		"type": opType,
	}
	var data AuditData
	n, err := mdl.QueryBuilder().Where(cond).OrderBy("id", "desc").First(&data)
	if err != nil {
		return nil, 0, err
	}
	return &data, n, nil
}

func (mdl *AuditModel) GetByRelateID(opType, relateID string) ([]AuditData, int, error) {
	cond := map[string]any{
		"relate_id": relateID,
		"type":      opType,
	}
	var data []AuditData
	n, err := mdl.QueryBuilder().Where(cond).OrderBy("id", "desc").Find(&data)
	if err != nil {
		return nil, 0, err
	}
	return data, n, nil
}
