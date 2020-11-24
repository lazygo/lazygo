package lazygo

import (
	"github.com/lazygo/lazygo/utils"
	"net/http"
	"reflect"
	"strings"
)

type Router struct {
	mux *http.ServeMux
}

// 实例化路由器，在框架初始化时调用一次
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// 注册控制器
// 注册后会按照请求路径自动路由到控制器的方法
func (r *Router)RegisterController(c IController) {
	// 反射解析控制器类型及路径信息
	v := reflect.ValueOf(c)
	ct := reflect.Indirect(v).Type()
	name := ct.Name()
	pkg := strings.SplitN(ct.PkgPath(), "controller", 2)
	path := "/"
	if len(pkg) >= 2 && pkg[1] != "" {
		path = path + strings.Trim(pkg[1], "/") + "/"
	}
	path = path + strings.ToLower(name[:len(name)-10]) + "/"

	//
	h := func(resp http.ResponseWriter, req *http.Request) {
		uri := req.URL.Path
		action := strings.TrimPrefix(uri, path)
		if action == "" {
			action = "Index"
		}
		action = utils.ToCamelString(action)
		// 通过反射实例化控制器
		cv := reflect.New(ct)
		ctl := cv.Interface().(IController)
		// 为控制器对象设置请求及相应
		ctl.setContext(resp, req)
		ctl.Init()
		m := cv.MethodByName(action + "Action")
		// panic处理
		defer ctl.HandleError()
		if !m.IsValid() {
			panic(ErrNotFound)
		}
		// 执行路由方法
		m.Interface().(func())()
	}
	r.mux.HandleFunc(path, h)
}

// 获取路由器，用于为server提供mux，server初始化时调用
func (r *Router) GetHandle() *http.ServeMux {
	return r.mux
}