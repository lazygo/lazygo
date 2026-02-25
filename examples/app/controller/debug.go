package controller

import (
	"fmt"
	"io"

	request "github.com/lazygo-dev/lazygo/examples/app/request/debug"
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo/lazygo/server"
)

type DebugController struct {
	Ctx framework.Context
}

// 获取客户端IP
func (ctl *DebugController) ClientIP() error {
	var content string

	content += fmt.Sprintf("%v\n", ctl.Ctx.RealIP())
	content += fmt.Sprintf("%v\n", ctl.Ctx.Request().Header.Get("X-Forwarded-For"))
	content += fmt.Sprintf("%#v\n", ctl.Ctx.Request().Header)

	return ctl.Ctx.Blob(200, server.MIMETextPlain, []byte(content))
}

// 获取请求体
func (ctl *DebugController) DumpPostBody(req *request.DumpPostBodyRequest) error {
	var content string

	content += fmt.Sprintf("%v\n", req)
	data, _ := io.ReadAll(ctl.Ctx.Request().Body)
	content += fmt.Sprintf("%s\n", string(data))

	return ctl.Ctx.Blob(200, server.MIMETextPlain, []byte(content))
}
