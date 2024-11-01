package server

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lazygo/lazygo/utils"
)

// Controller 转为 server.HandlerFunc
func Controller(h any, methodName ...string) HandlerFunc {
	rtServ, serviceName, err := routes.Make(h)
	if err != nil {
		panic(err)
	}
	var name string
	if len(methodName) > 0 {
		name = utils.ToSnakeString(methodName[0])
	}

	return func(ctx Context) error {
		if name == "" {
			routePath := strings.TrimRight(ctx.GetRoutePath(), "/")
			index := strings.LastIndex(routePath, "/")
			name = strings.TrimLeft(routePath[index:], "/")
		}

		method, ok := routes[serviceName][name]
		if !ok {
			return ErrNotFound.SetInternal(fmt.Errorf(" method name %s not found", name))
		}
		rCtx := reflect.ValueOf(ctx)

		pServ := reflect.New(rtServ)
		args := []reflect.Value{pServ}
		var req any

		if method.Request != nil {
			pReq := reflect.New(method.Request)
			req = pReq.Interface()

			defer pReq.MethodByName("Clear").Call(nil)

			if err = ctx.Bind(req); err != nil {
				return ErrBadRequest.SetInternal(fmt.Errorf("params error, req: %v, err: %v", req, err))
			}
			verify := pReq.MethodByName("Verify")
			var params []reflect.Value
			if verify.Type().NumIn() > 0 {
				params = append(params, rCtx)
			}
			err := verify.Call(params)[0].Interface()
			if err != nil {
				if he, ok := err.(*HTTPError); ok {
					return he.SetInternal(fmt.Errorf("verify params fail, req: %v", req))
				}
				return ErrBadRequest.SetInternal(fmt.Errorf("params error, req: %v, err: %v", req, err))
			}
			args = append(args, pReq)
		}

		pServ.Elem().FieldByName("Ctx").Set(rCtx)

		out := method.Method.Func.Call(args)
		numOut := len(out)
		if numOut == 1 {
			if ierr := out[0].Interface(); ierr != nil {
				if err = ierr.(error); err != nil {
					if he, ok := err.(*HTTPError); ok {
						return he.SetInternal(fmt.Errorf("request fail, req: %v", req))

					}
					return fmt.Errorf("request fail, req: %v, err: %v", req, err)
				}
			}
			return nil
		}
		if numOut == 2 {
			resp := out[0].Interface()
			if ierr := out[1].Interface(); ierr != nil {
				if err = ierr.(error); err != nil {
					if he, ok := err.(*HTTPError); ok {
						return he.SetInternal(fmt.Errorf("request fail, req: %v, resp: %v", req, resp))

					}
					return fmt.Errorf("request fail, req: %v, resp: %v, err: %v", req, resp, err)
				}
			}
			return ctx.s().HTTPOKHandler(resp, ctx)
		}
		return ErrInternalServerError
	}
}
