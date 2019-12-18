package lazygo

import (
	"github.com/gorilla/mux"
	"github.com/lazygo/lazygo/library"
	"net/http"
	"reflect"
	"strings"
)

type Router struct {
	router  *mux.Router
	actions map[string]map[string]bool
}

func NewRouter() (*Router, error) {
	router := mux.NewRouter()
	return &Router{
		router:  router,
		actions: map[string]map[string]bool{},
	}, nil
}

func (r *Router) HandleFunc(path string, c interface{}) {
	r.router.HandleFunc(path, r.Handler(c))
}

func (r *Router) Handler(c interface{}) func(res http.ResponseWriter, req *http.Request) {
	v := reflect.ValueOf(c)
	vt := reflect.TypeOf(c)
	ct := reflect.Indirect(v).Type()
	name := ct.Name()

	// 获取控制器方法列表
	if _, ok := r.actions[name]; !ok {
		methodNum := vt.NumMethod()
		r.actions[name] = map[string]bool{}
		// 获取方法
		for i := 0; i < methodNum; i++ {
			m := vt.Method(i)
			methodName := m.Name
			if strings.HasSuffix(methodName, "Action") {
				methodName = methodName[0 : len(methodName)-6]
				r.actions[name][methodName] = true
			}
		}
	}

	h := func(res http.ResponseWriter, req *http.Request) {
		action := mux.Vars(req)["action"]
		if action == "" {
			action = "Index"
		}
		action = library.ToCamelString(action)
		c := reflect.New(ct)
		base := c.Interface().(IController)
		base.Init(res, req)

		if _, ok := r.actions[name][action]; !ok {
			base.Abort(404, "404 page not found")
			return
		}
		m := c.MethodByName(action + "Action")
		if !m.IsValid() {
			base.Abort(404, "404 page not found")
			return
		}
		a := m.Interface().(func())
		base.Run(a)
	}
	return h
}

func (r *Router) HandleStatic(asset func(string) ([]byte, error)) {
	r.router.HandleFunc("/{path:static/.+}", func(res http.ResponseWriter, req *http.Request) {
		path := mux.Vars(req)["path"]
		data, err := asset(path)
		if err != nil {
			res.WriteHeader(404)
			data = []byte("404 page not found")
		}
		mimetype := library.GetMimeType(path)
		res.Header().Set("Content-Type", mimetype)
		res.Write(data)
	})
}

func (r *Router) GetHandle() *mux.Router {
	return r.router
}
