package middleware

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils/errors"
)

// AuthUser 用户端登录校验
func AuthUser(next server.HandlerFunc) server.HandlerFunc {
	return framework.ToBase(func(ctx framework.Context) error {
		if ctx.GetUID() == 0 {
			return errors.ErrUnauthorized
		}
		return next(ctx)
	})
}
