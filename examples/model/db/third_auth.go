package db

import (
	"time"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/model"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
)

var ThirdMap = map[string]cacheModel.AuthUserData{
	"5aee2b10e5758e2bb7adfac13a31827b": {UID: 24, Appid: AppidMain}, // 测试
}

var (
	AppidSecret = map[int]string{
		10001: "dd89240a71935f66f460908c7b66af87", // test
	}
)

type ThirdAuthModel struct {
	Ctx framework.Context
	model.TxModel[ThirdAuthData]
}

type ThirdAuthData struct {
	UID      uint64 `json:"uid"`
	Appid    string `json:"appid"`
	ThirdUID string `json:"third_uid"`
	Vendor   string `json:"vendor"`
	Data     string `json:"data"`
	CTime    int64  `json:"ctime"`
	MTime    int64  `json:"mtime"`
}

func NewThirdAuthModel(ctx framework.Context) *ThirdAuthModel {
	mdl := &ThirdAuthModel{Ctx: ctx}
	mdl.SetTable("uc_third_auth")
	mdl.SetDB("lazygo-db")
	return mdl
}

func (mdl *ThirdAuthModel) Create(data map[string]any) (int64, error) {
	data["ctime"] = time.Now().Unix()
	return mdl.QueryBuilder().Insert(data)
}

func (mdl *ThirdAuthModel) UpdateByUid(uid int64, data map[string]any) (int64, error) {
	cond := map[string]any{
		"uid": uid,
	}
	data["mtime"] = time.Now().Unix()
	return mdl.QueryBuilder().Where(cond).Update(data)
}
