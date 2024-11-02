package server

import (
	"errors"
	"fmt"
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
		return nil, "", errors.New("handler must be a controller")
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

			numIn := method.Type.NumIn()
			switch numIn {
			case 2:
				if _, ok := method.Type.In(1).MethodByName("Clear"); !ok {
					return nil, "", fmt.Errorf("method %s args not implement Request, need func Clear()", methodName)
				}
				rf, ok := method.Type.In(1).MethodByName("Verify")
				if !ok {
					return nil, "", fmt.Errorf("method %s args not implement Request, need func Verify(Context) error", methodName)
				}
				if rf.Type.NumOut() != 1 || !rf.Type.Out(0).Implements(tError) {
					return nil, "", fmt.Errorf("method %s args not implement Request, need func Verify(Context) error", methodName)
				}
				fallthrough
			case 1:
				if method.Type.In(0).String() != serviceName {
					return nil, "", fmt.Errorf("method %s must has a receiver", methodName)
				}
			default:
				return nil, "", fmt.Errorf("method %s param num error", methodName)
			}

			numOut := method.Type.NumOut()
			if numOut > 2 || numOut < 1 {
				return nil, "", fmt.Errorf("method %s must 1 or 2 return args", methodName)
			}
			if method.Type.Out(numOut-1).Name() != "error" {
				return nil, "", fmt.Errorf("method %s last return args must implement error", methodName)
			}

			r := Route{Method: method}
			if numIn == 2 {
				r.Request = method.Type.In(1).Elem()
			}
			routes[serviceName][methodName] = r
		}
	}
	return rt, serviceName, nil
}
