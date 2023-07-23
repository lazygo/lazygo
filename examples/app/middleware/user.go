package middleware

import (
	"github.com/lazygo/lazygo/examples/framework"
	cacheModel "github.com/lazygo/lazygo/examples/model/cache"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
)

// User 解析用户token，设置uid
func User(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		// 检查此次请求中是否带有 token，如果没有则抛出异常。
		token := ctx.GetRequestHeader(server.HeaderAuthorization)

		cache := cacheModel.NewAuthUserCache()
		auth, _, err := cache.Get(token)
		if err != nil {
			ctx.Logger().Warn("[msg: get auth cache fail] [error: db error] [err: %v]", err)
			return errors.ErrInternalServerError
		}

		// WithValue 设置uid，可使用ctx.GetUID()取出
		ctx.WithValue("uid", auth.UID)

		ctx.Logger().Info("[msg: auth info] [auth: %v]", auth)

		return next(ctx)
	})
}
