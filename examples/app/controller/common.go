package controller

import (
	"context"
	"fmt"
	"hash/crc32"
	"path"
	"strings"
	"time"

	"github.com/coder/websocket"
	request "github.com/lazygo/lazygo/examples/app/request/common"
	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/pkg/cos"
)

const (
	UploadDomain = "https://img.domain.com/"
)

type CommonController struct {
	Ctx framework.Context
}

func (s *CommonController) Upload(req *request.ToolsUploadRequest) (any, error) {
	uid := s.Ctx.UID()
	filename := fmt.Sprintf("%d%d", uid, time.Now().UnixNano()/1000) + path.Ext(req.Image.FileHeader.Filename)

	// save path
	sum := crc32.ChecksumIEEE([]byte(filename))
	dir := fmt.Sprintf("upload/%02x/%02x/", sum>>24&0xff, sum>>16&0xff)
	filePath := path.Join(dir, filename)
	err := cos.Upload(context.Background(), filePath, req.Image.File)
	if err != nil {
		s.Ctx.Logger().Warn("[msg: 上传文件失败] [err: %v]", err)
		return nil, errors.ErrUploadCosFail
	}

	data := map[string]any{
		"uid":         uid,
		"uri":         filePath,
		"name":        filename,
		"origin_name": req.Image.FileHeader.Filename,
		"size":        req.Image.FileHeader.Size,
		"category":    req.Category,
		"type":        dbModel.UploadKindImage,
		"ctime":       time.Now().Unix(),
	}
	mdlUpload := dbModel.NewUploadModel(s.Ctx)
	id, err := mdlUpload.QueryBuilder().Insert(data)
	if err != nil {
		s.Ctx.Logger().Warn("[msg: 上传文件保存失败] [err: %v]", err)
		return nil, errors.ErrDBError
	}

	resp := &request.ToolsUploadResponse{
		ID:  id,
		URL: fmt.Sprintf("%s/%s", strings.TrimRight(UploadDomain, "/"), strings.TrimLeft(filePath, "/")),
	}
	return resp, nil
}

// Connection 连接websocket
// Websocket 请求格式：{
// "id": 1, // 唯一消息ID，整数，必填，返回时可通过此ID关联响应
// "uri": "path/to/uri?query=value", // 请求uri，必填
// "header": { // 请求头，可选
// "Key1": "Value1",
// "Key2": "Value2",
// ...
// },
// "body": "xxxxxxx" // 请求体，可选
// }
func (ctl *CommonController) Connection(req *request.ConnectionRequest) error {
	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
		//Subprotocols:    []string{"lazygo"},
		CompressionMode: websocket.CompressionContextTakeover,
	}
	w := ctl.Ctx.ResponseWriter()
	r := ctl.Ctx.Request()
	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		ctl.Ctx.Logger().Warn("[msg: accept websocket connection failed] [err: %v]", err)
		return fmt.Errorf("accept websocket connection failed: %w", err)
	}

	sub := fmt.Sprintf("uid:%d", req.UID)
	return ctl.Ctx.Event(server.MethodWebSocket, sub).Serve(ctl.Ctx, ctl.Ctx.RequestID(), server.WebSocketWrapper(ctl.Ctx, conn))
}
