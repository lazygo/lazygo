package framework

import (
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/lazygo/utils"
)

type HandlerFunc func(Context) error
type HTTPErrorHandler func(error, Context)

// ToBase HandlerFunc 转为 server.HandlerFunc
func ToBase(h HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc := c.(*context)
		return h(cc)
	}
}

// HandleError 返回失败
func ToBaseHTTPErrorHandler(h HTTPErrorHandler) server.HTTPErrorHandler {
	return func(err error, c server.Context) {
		cc := c.(*context)
		h(err, cc)
	}
}

func GenTraceID() uint64 {
	var x = strconv.Itoa(time.Now().Nanosecond() / 1000)
	res, errParseInt := strconv.ParseInt(x, 10, 64)
	if errParseInt != nil {
		return 0
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := ((time.Now().Unix()*100000+res)&0xFFFFFFFF)*1000000000 + 100000000 + r.Int63n(899999999)
	return uint64(id)
}

// ExtendContextMiddleware Context 拓展中间件
func ExtendContextMiddleware(h server.HandlerFunc) server.HandlerFunc {
	return func(c server.Context) error {
		cc := &context{c}
		return h(cc)
	}
}

// HandleSucc 返回成功
func HandleSucc() server.HandlerFunc {
	return ToBase(func(ctx Context) error {
		return ctx.Succ(struct{}{})
	})
}

// Controller 转为 server.HandlerFunc
func Controller(h interface{}, methodName ...string) server.HandlerFunc {
	rtServ, serviceName, err := routes.Make(h)
	if err != nil {
		panic(err)
	}
	var alias string
	if len(methodName) > 0 {
		alias = utils.ToSnakeString(methodName[0])
	}

	return ToBase(func(ctx Context) error {
		name := alias
		if name == "" {
			routePath := strings.TrimRight(ctx.GetRoutePath(), "/")
			index := strings.LastIndex(routePath, "/")
			name = strings.TrimLeft(routePath[index:], "/")
		}

		method, ok := routes[serviceName][name]
		if !ok {
			ctx.Logger().Warn("[msg: not fount] [method name: %s]", name)
			return server.ErrNotFound
		}
		pReq := reflect.New(method.Request)
		req := pReq.Interface().(Request)

		defer req.Clear()

		if err = ctx.Bind(req); err != nil {
			ctx.Logger().Warn("[msg: params error] [req: %v] [err: %v]", req, err)
			return server.ErrBadRequest
		}
		if err = req.Verify(); err != nil {
			ctx.Logger().Warn("[msg: verify params fail] [resp: %v] [err: %v]", req, err)
			return err
		}

		pServ := reflect.New(rtServ)
		pServ.Elem().FieldByName("Ctx").Set(reflect.ValueOf(ctx))

		out := method.Method.Func.Call([]reflect.Value{pServ, pReq})
		numOut := len(out)
		if numOut == 1 {
			if ierr := out[0].Interface(); ierr != nil {
				if err = ierr.(error); err != nil {
					ctx.Logger().Warn("[msg: request fail] [req: %v] [err: %v]", req, err)
					return err
				}
			}
			return nil
		}
		if numOut == 2 {
			resp := out[0].Interface()
			if ierr := out[1].Interface(); ierr != nil {
				if err = ierr.(error); err != nil {
					ctx.Logger().Warn("[msg: request fail] [req: %v] [resp: %v] [err: %v]", req, resp, err)
					return err
				}
			}
			return ctx.Succ(resp)
		}
		ctx.Logger().Warn("[msg: method return value error] [out: %v]", out)
		return server.ErrInternalServerError
	})
}
