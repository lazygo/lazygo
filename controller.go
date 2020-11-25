package lazygo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lazygo/lazygo/utils"
	"io/ioutil"
	"net/http"
	"sync"
)

var (
	ErrNotFound     = errors.New("404 page not found") // 404错误
	HttpResponseEnd = 0                                // Http响应完成
)

type IController interface {
	setContext(res http.ResponseWriter, req *http.Request)
	Init()
	HandleError()
}

// 控制器
type Controller struct {
	Res        http.ResponseWriter
	Req        *http.Request
	tpl        *Template
	errHandler sync.Map
}

// 设置请求及响应
func (c *Controller) setContext(res http.ResponseWriter, req *http.Request) {
	c.Res = res
	c.Req = req
}

// 控制器初始化时调用
// 可以控制器中，重写此方法
func (c *Controller) Init() {
	c.InitTpl("templates/", ".html", nil)
	c.ErrorHandler(ErrNotFound, func(err error) {
		c.Abort(404, err.Error())
	})
}

// 初始化模板
// 可以控制器中，重写此方法
func (c *Controller) InitTpl(prefix string, suffix string, asset AssetRegister) {
	c.tpl = NewTemplate(c.Res, c.Req, prefix, suffix, asset)
	c.ErrorHandler(ErrNotFound, func(err error) {
		c.Abort(404, err.Error())
	})
}

// 注册错误对应的处理
func (c *Controller) ErrorHandler(err error, handleError func(err error)) {
	c.errHandler.Store(err, handleError)
}

// Error处理
// 用于recover代码中的panic
func (c *Controller) HandleError() {
	defer func() { recover() }()
	e := recover()
	if e == nil {
		return
	}
	if e == HttpResponseEnd {
		return
	}
	err, ok := e.(error)
	if !ok {
		c.Abort(500, fmt.Sprintln(err))
	}

	// 找到了指定的错误处理
	if handle, ok := c.errHandler.Load(err); ok {
		if handle != nil {
			handle.(func(err error))(err)
		}
		return
	}
	c.Abort(500, err.Error())
}

// GetBody 获取原始请求body
// 将io中的请求body全部读入到内存，并将Req.Body指向该内存
// 注意：如果上传大文件，调用此方法会占用大量内存
func (c *Controller) GetBody() ([]byte, error) {
	buf, err := ioutil.ReadAll(c.Req.Body)
	if err != nil {
		return nil, err
	}
	c.Req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	return buf, nil
}

// 获取Get字符串变量
func (c *Controller) GetString(name string) string {
	return utils.ToString(c.Req.URL.Query().Get(name))
}

// 获取Get整型变量
func (c *Controller) GetInt(name string) int {
	return utils.ToInt(c.Req.URL.Query().Get(name))
}

// 获取Post字符串变量
func (c *Controller) PostString(name string) string {
	return utils.ToString(c.Req.FormValue(name))
}

// 获取Post整型变量
func (c *Controller) PostInt(name string) int {
	return utils.ToInt(c.Req.FormValue(name))
}

// 终止当前请求执行 并写入http响应
func (c *Controller) Abort(httpCode int, message string) {
	c.Res.WriteHeader(httpCode)
	c.Res.Write([]byte(message))
	// 执行Response后当前请求的代码就不再往后执行了
	panic(HttpResponseEnd)
}

func (c *Controller) SetHeader(headerOptions map[string]string) *Controller {
	if len(headerOptions) > 0 {
		for field, val := range headerOptions {
			c.Res.Header().Set(field, val)
		}
	}
	return c
}

// 成功响应
func (c *Controller) ApiSucc(data map[string]interface{}, message string) {
	if data == nil {
		data = map[string]interface{}{}
	}
	result := map[string]interface{}{
		"code":    200,
		"message": message,
		"data":    data,
	}
	c.JsonResponse(result)
}

// 失败响应
func (c *Controller) ApiFail(code int, message string, data interface{}) {
	result := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	c.JsonResponse(result)
}

// json响应
func (c *Controller) JsonResponse(data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		jsonData = nil
	}
	c.Res.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonData = append(jsonData, '\n')
	c.Res.Write(jsonData)
	// 执行Response后当前请求的代码就不再往后执行了, 用于防止漏掉return语句 导致返回多个连在一起的json
	panic(HttpResponseEnd)
}

func (c *Controller) AssignMap(data map[string]interface{}) *Template {
	for key, val := range data {
		c.tpl.TplData[key] = val
	}
	return c.tpl
}

func (c *Controller) Assign(key string, data interface{}) *Template {
	c.tpl.TplData[key] = data
	return c.tpl
}

func (c *Controller) Display(tpl string) {
	c.tpl.Display(tpl)
}
