package debug

import (
	"github.com/lazygo/lazygo/examples/framework"
)

type DumpPostBodyRequest struct {
	Appid int `json:"appid" bind:"context"`
	// 此处可以直接通过绑定user middleware中WithValue方法设置的uid
	TestField1 string `json:"test_field_1" bind:"form" process:"trim,tolower"`
}

func (r *DumpPostBodyRequest) Verify(ctx framework.Context) error {

	return nil
}

func (r *DumpPostBodyRequest) Clear() {
}
