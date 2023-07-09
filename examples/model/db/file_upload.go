package db

import (
	"github.com/lazygo/lazygo/examples/model"
)

type UploadModel struct {
	model.DbModel
}

const (
	UploadKindDefault = iota
	UploadKindImage
)

type UploadData struct {
	Id uint64 `json:"id"`
}

func NewUploadModel() *UploadModel {
	mdl := &UploadModel{}
	mdl.SetTable("file_upload")
	mdl.SetDb("hd")
	return mdl
}
