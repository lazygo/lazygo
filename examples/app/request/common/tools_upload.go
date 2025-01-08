package request

import (
	"path"

	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
	"golang.org/x/exp/slices"
)

type ToolsUploadRequest struct {
	Category string      `json:"category" bind:"query,form" process:"trim,cut(20)"`
	Image    server.File `json:"image"`
}

type ToolsUploadResponse struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

func (r *ToolsUploadRequest) Verify() error {

	if slices.Contains(utils.ImageFormat, path.Ext(r.Image.FileHeader.Filename)) == false {
		return errors.ErrInvalidImageFormat
	}
	return nil
}

func (r *ToolsUploadRequest) Clear() {
	if r.Image.File != nil {
		r.Image.File.Close()
	}
}
