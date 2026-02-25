package middleware

import (
	"github.com/lazygo/lazygo/server"

	"github.com/lazygo/lazygo/examples/framework"
)

// DEBUG
func Debug(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		ctx.ResponseWriter().Header().Set("Access-Control-Allow-Origin", "*")
		ctx.ResponseWriter().Header().Set("Access-Control-Allow-Credentials", "true")

		return next(ctx)
	})
}
