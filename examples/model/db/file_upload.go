package db

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/model"
)

type UploadModel struct {
	Ctx framework.Context
	model.TxModel[UploadData]
}

const (
	UploadKindDefault = iota
	UploadKindImage
)

type UploadData struct {
	Id uint64 `json:"id"`
}

func NewUploadModel(ctx framework.Context) *UploadModel {
	mdl := &UploadModel{Ctx: ctx}
	mdl.SetTable("file_upload")
	mdl.SetDB("lazygo-db")
	return mdl
}
