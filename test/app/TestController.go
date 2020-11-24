package app

import (
	"fmt"
	"github.com/lazygo/lazygo/utils"
)

type TestController struct {
	Controller
}

//
func (t *TestController) Init() {
	t.Controller.Init()
}

func (t *TestController) TestResponseAction() {
	t.ApiSucc(nil, "^_^")
}

func (t *TestController) TestEndAction() {
	fmt.Println(t.GetBody())
	fmt.Println(t.GetString("aa"))
	fmt.Println(t.PostString("aa"))
	t.ApiFail(10, "┭┮﹏┭┮",  map[string]interface{}{"a":1})
	t.ApiFail(22, "┭┮",  nil)
}

func (t *TestController) TestErrorAction() {
	fmt.Println(t.GetString("aa"))
	fmt.Println(t.PostString("aa"))
	panic(err)
	t.ApiFail(10, "┭┮﹏┭┮",  map[string]interface{}{"a":1})
}

func (t *TestController) TestError2Action() {
	panic(err2)
	t.ApiFail(11, "(_)",  "")
}

func (t *TestController) TestMimeAction() {
	mimetype := utils.GetMimeType("/path/to/aaa.html")
	t.ApiSucc(nil,  mimetype)
}