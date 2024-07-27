package server

import (
	"errors"
	"reflect"

	"github.com/lazygo/lazygo/utils"
)

type Route struct {
	Method  reflect.Method
	Request reflect.Type
}

type RouteCache map[string]map[string]Route

var routes = make(RouteCache)

func (RouteCache) Make(h any) (reflect.Type, string, error) {
	rv := reflect.Indirect(reflect.ValueOf(h))
	if rv.Kind() != reflect.Struct {
		return nil, "", errors.New("param not service")
	}
	rt := rv.Type()
	rtp := reflect.PointerTo(rt)
	serviceName := rtp.String()
	num := rtp.NumMethod()
	if _, ok := routes[serviceName]; !ok {
		routes[serviceName] = make(map[string]Route, num)
		tError := reflect.TypeOf((*error)(nil)).Elem()
		for i := 0; i < num; i++ {
			method := rtp.Method(i)
			methodName := utils.ToSnakeString(method.Name)
			if method.Type.NumIn() != 2 {
				return nil, "", errors.New("method param num error")
			}

			if method.Type.In(0).String() != serviceName {
				return nil, "", errors.New("method first param must type receiver")
			}
			if _, ok := method.Type.In(1).MethodByName("Clear"); !ok {
				return nil, "", errors.New("method second param Request must has func Clear()")
			}
			rf, ok := method.Type.In(1).MethodByName("Verify")
			if !ok {
				return nil, "", errors.New("method second param Request must has func Verify(Context) error")
			}
			if rf.Type.NumOut() != 1 || !rf.Type.Out(0).Implements(tError) {
				return nil, "", errors.New("method second param Request must has func Verify(Context) error")
			}

			numOut := method.Type.NumOut()
			if numOut > 2 || numOut < 1 {
				return nil, "", errors.New("method return num error")
			}
			if method.Type.Out(numOut-1).Name() != "error" {
				return nil, "", errors.New("method second return must type error")
			}

			routes[serviceName][methodName] = Route{
				Method:  method,
				Request: method.Type.In(1).Elem(),
			}
		}
	}
	return rt, serviceName, nil
}
