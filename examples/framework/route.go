package framework

import (
	"errors"
	"reflect"

	"github.com/lazygo/lazygo/utils"
)

type Request interface {
	Verify() error
	Clear()
}

type Route struct {
	Method  reflect.Method
	Request reflect.Type
}

type RouteCache map[string]map[string]Route

var routes = make(RouteCache)

func (RouteCache) Make(h interface{}) (reflect.Type, string, error) {
	rv := reflect.Indirect(reflect.ValueOf(h))
	if rv.Kind() != reflect.Struct {
		return nil, "", errors.New("param not service")
	}
	rt := rv.Type()
	rtp := reflect.PtrTo(rt)
	serviceName := rtp.String()
	num := rtp.NumMethod()
	if _, ok := routes[serviceName]; !ok {
		routes[serviceName] = make(map[string]Route, num)
		for i := 0; i < num; i++ {
			method := rtp.Method(i)
			methodName := utils.ToSnakeString(method.Name)
			if method.Type.NumIn() != 2 {
				return nil, "", errors.New("method param num error")
			}

			if method.Type.In(0).String() != serviceName {
				return nil, "", errors.New("method first param must type receiver")
			}
			if !method.Type.In(1).Implements(reflect.TypeOf((*Request)(nil)).Elem()) {
				return nil, "", errors.New("method second param must implements Request")
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
