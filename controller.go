package lazygo

import (
	"encoding/json"
	"fmt"
	"github.com/lazygo/lazygo/utils"
	"net/http"
	"runtime/debug"
)

type IController interface {
	Init(res http.ResponseWriter, req *http.Request)
	Run(action func())
	Abort(httpCode int, message string)
}

type Controller struct {
	Res http.ResponseWriter
	Req *http.Request
	tpl *Template
}

func (c *Controller) Init(res http.ResponseWriter, req *http.Request) {
	c.Res = res
	c.Req = req
	c.tpl = NewTemplate(res, req, "templates/", ".html")
}

func (c *Controller) GetString(name string) string {
	return utils.ToString(c.Req.URL.Query().Get(name))
}

func (c *Controller) GetInt(name string) int {
	return utils.ToInt(c.Req.URL.Query().Get(name))
}

func (c *Controller) PostString(name string) string {
	return utils.ToString(c.Req.FormValue(name))
}

func (c *Controller) PostInt(name string) int {
	return utils.ToInt(c.Req.FormValue(name))
}

func (c *Controller) Run(action func()) {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		if err := recover(); err != nil {
			fmt.Println(err)                   // 这里的err其实就是panic传入的内容
			fmt.Println(string(debug.Stack())) // 这里的err其实就是panic传入的内容
			c.Abort(500, err.(error).Error())
		}
	}()
	action()
}

func (c *Controller) Abort(httpCode int, message string) {
	c.Res.WriteHeader(httpCode)
	c.Res.Write([]byte(message))
}

func (c *Controller) SetHeader(headerOptions map[string]string) *Controller {
	if len(headerOptions) > 0 {
		for field, val := range headerOptions {
			c.Res.Header().Set(field, val)
		}
	}
	return c
}

func (c *Controller) Response(code int, message string, data map[string]interface{}) {
	result := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		jsonData = nil
	}
	c.Res.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Res.Write(jsonData)
}

func (c *Controller) AssignMap(data map[string]interface{}) *Template {
	return c.AssignMap(data)
}

func (c *Controller) Assign(key string, data interface{}) *Template {
	return c.Assign(key, data)
}

func (c *Controller) Display(tpl string) {
	c.tpl.Display(tpl)
}
