package middleware

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/server"
)

// RequestID 添加request_id
func RequestID(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {

		var rid uint64
		if param := ctx.RequestHeader(server.HeaderXRequestID); param != "" {
			var err error
			rid, err = strconv.ParseUint(param, 10, 64)
			if err != nil {
				ctx.Logger().Warn(
					"[msg: request id error] [%s: %s] [err: %v]",
					server.HeaderXRequestID, param, err,
				)
			}
		}
		if rid == 0 {
			rid = generator()
		}

		ctx.WithValue(server.HeaderXRequestID, rid)
		ctx.ResponseWriter().Header().Set(server.HeaderXRequestID, strconv.FormatUint(rid, 10))

		return next(ctx)
	})
}

func generator() uint64 {
	var x = strconv.Itoa(time.Now().Nanosecond() / 1000)
	res, errParseInt := strconv.ParseInt(x, 10, 64)
	if errParseInt != nil {
		return 0
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := ((time.Now().Unix()*100000+res)&0xFFFFFFFF)*1000000000 + 100000000 + r.Int63n(899999999)
	return uint64(id)
}
