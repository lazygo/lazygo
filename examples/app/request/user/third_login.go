package request

import (
	"github.com/lazygo/lazygo/examples/framework"
)

// 设备登录时复用此方法，即支持json请求也支持来自设备的formdata请求
type ThirdLoginRequest struct {
	Appid      int            `json:"appid" bind:"query,form" process:"trim"`
	ThirdUID   string         `json:"third_uid" bind:"query,form" process:"trim"`
	Vendor     string         `json:"vendor" bind:"query,form" process:"trim"`
	AutoCreate int            `json:"auth_create" bind:"query,form"`          // 如果账号不存在则自动创建账号
	Extra      map[string]any `json:"extra" bind:"query,form" process:"trim"` // 其他额外信息
}

type ThirdLoginResponse struct {
	Token string         `json:"token"`
	URL   string         `json:"url"`
	Extra map[string]any `json:"extra"`
}

func (r *ThirdLoginRequest) Verify(ctx framework.Context) error {

	return nil
}

func (r *ThirdLoginRequest) Clear() {
}
