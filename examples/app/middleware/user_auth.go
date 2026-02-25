package middleware

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
)

// AuthUser 用户端登录校验
func AuthUser(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		if ctx.UID() == 0 {
			return errors.ErrUnauthorized
		}
		return next(ctx)
	})
}
