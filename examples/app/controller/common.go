package controller

import (
	"fmt"
	"hash/crc32"
	"path"
	"time"

	request "github.com/lazygo/lazygo/examples/app/request/common"
	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/examples/pkg/cos"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

const (
	UploadDomain = "https://img.domain.com/"
)

type CommonController struct {
	Ctx framework.Context
}

func (s *CommonController) Upload(req *request.ToolsUploadRequest) (*request.ToolsUploadResponse, error) {
	uid := s.Ctx.GetUID()
	filename := fmt.Sprintf("%d%d", uid, time.Now().UnixNano()/1000) + path.Ext(req.Image.FileHeader.Filename)

	// crc path
	sum := crc32.ChecksumIEEE([]byte(filename))
	dir := fmt.Sprintf("upload/%02x/%02x/", sum>>24&0xff, sum>>16&0xff)
	err := cos.Upload("file", "req.Image.File")
	if err != nil {
		s.Ctx.Logger().Warn("[msg: Cos 上传文件失败] [err: %v]", err)
		return nil, errors.ErrUploadCosFail
	}
	uri := dir + filename

	data := map[string]interface{}{
		"uid":         uid,
		"uri":         uri,
		"name":        filename,
		"origin_name": req.Image.FileHeader.Filename,
		"size":        req.Image.FileHeader.Size,
		"category":    req.Category,
		"type":        dbModel.UploadKindImage,
		"ctime":       time.Now().Unix(),
	}
	mdlUpload := dbModel.NewUploadModel()
	id, err := mdlUpload.QueryBuilder().Insert(data)
	if err != nil {
		s.Ctx.Logger().Warn("[msg: 上传文件保存失败] [err: %v]", err)
		return nil, errors.ErrDbError
	}

	resp := &request.ToolsUploadResponse{
		ID:  id,
		URL: UploadDomain + uri,
	}
	return resp, nil
}
