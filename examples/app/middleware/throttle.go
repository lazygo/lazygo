package middleware

import (
	"time"

	"github.com/lazygo-dev/lazygo/examples/framework"
	cacheModel "github.com/lazygo-dev/lazygo/examples/model/cache"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
)

// Throttle 节流阀，保证一个路由内duration时间内最多执行n次
func Throttle(n int64, duration time.Duration) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return framework.BaseHandlerFunc(func(ctx framework.Context) error {
			cacheTicker := cacheModel.NewTickerCache(ctx)
			ok, err := cacheTicker.Can(ctx.GetRoutePath(), n, duration)
			if err != nil {
				ctx.Logger().Error("[msg: ticker cache error] [error: db error] [err: %v]", err)
				return errors.ErrDBError
			}
			if !ok {
				ctx.Logger().Info("[msg: too many attempts] [err: %v]", err)
				return errors.ErrTooManyAttempts
			}
			ctx.WithValue(cacheModel.ThrottleDuration, duration)
			ctx.WithValue(cacheModel.ThrottleNum, n)
			return next(ctx)
		})
	}

}
