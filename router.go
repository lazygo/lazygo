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

func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

func (r *Router)RegisterController(c IController) {
	v := reflect.ValueOf(c)
	ct := reflect.Indirect(v).Type()
	name := ct.Name()
	pkg := strings.SplitN(ct.PkgPath(), "/", 2)
	path := "/"
	if len(pkg) >= 2 {
		path = path + pkg[1] + "/"
	}
	path = path + strings.ToLower(name[:len(name)-10]) + "/"

	h := func(resp http.ResponseWriter, req *http.Request) {
		uri := req.URL.Path
		action := strings.TrimPrefix(uri, path)
		if action == "" {
			action = "Index"
		}
		action = utils.ToCamelString(action)
		ctl := reflect.New(ct)
		base := ctl.Interface().(IController)
		base.Init(resp, req)
		m := ctl.MethodByName(action + "Action")
		if !m.IsValid() {
			base.Abort(404, "404 page not found")
			return
		}
		a := m.Interface().(func())
		base.Run(a)
	}
	r.mux.HandleFunc(path, h)
}

func (r *Router) GetHandle() *http.ServeMux {
	return r.mux
}