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
		pReq := reflect.New(method.Request)
		req := pReq.Interface().(Request)

		defer req.Clear()

		if err = ctx.Bind(req); err != nil {
			return ErrBadRequest.SetInternal(fmt.Errorf("params error, req: %v, err: %v", req, err))
		}
		if err = req.Verify(); err != nil {
			if he, ok := err.(*HTTPError); ok {
				return he.SetInternal(fmt.Errorf("verify params fail, req: %v", req))
			}
			return ErrBadRequest.SetInternal(fmt.Errorf("params error, req: %v, err: %v", req, err))
		}

		pServ := reflect.New(rtServ)
		pServ.Elem().FieldByName("Ctx").Set(reflect.ValueOf(ctx))

		out := method.Method.Func.Call([]reflect.Value{pServ, pReq})
		numOut := len(out)
		if numOut == 1 {
			if ierr := out[0].Interface(); ierr != nil {
				if err = ierr.(error); err != nil {
					if he, ok := err.(*HTTPError); ok {
						return he.SetInternal(fmt.Errorf("request fail, req: %v", req))

					}
					return fmt.Errorf("request fail, req: %v, err: %w", req, err)
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
					return fmt.Errorf("request fail, req: %v, resp: %v, err: %w", req, resp, err)
				}
			}
			return ctx.s().HTTPOKHandler(resp, ctx)
		}
		return ErrInternalServerError
	}
}
