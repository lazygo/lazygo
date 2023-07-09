package middleware

import (
	"strings"

	"github.com/lazygo/lazygo/server"
)

// StripUrlSuffix 去除url后缀
func StripUrlSuffix(next server.HandlerFunc) server.HandlerFunc {
	return func(ctx server.Context) error {
		path := ctx.Request().URL.Path

		switch {
		case strings.HasSuffix(path, ".json"):
			path = strings.TrimSuffix(path, ".json")
		}

		ctx.Request().URL.Path = path

		return next(ctx)
	}
}
