package middleware

import (
	"strings"

	"github.com/lazygo-dev/lazygo/examples/framework"
	cacheModel "github.com/lazygo-dev/lazygo/examples/model/cache"
	"github.com/lazygo-dev/lazygo/examples/utils"
	"github.com/lazygo-dev/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/pkg/token/jwt"
)

// User 解析用户token，设置uid
func User(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		// 从请求中获取token,解析uid,如果没有传递token,uid设置为0
		token := ctx.RequestHeader(server.HeaderAuthorization)
		if token == "" {
			token, _ = ctx.Param("token")
		}
		token = strings.Trim(token, "'\"")
		token = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(token), "Bearer"))
		if token != "" {
			// 兼容jwt认证方式
			checkCode := func(code string) error {
				cache := cacheModel.NewAuthUserCache(ctx)
				auth, ok, err := cache.Get(code)
				if err != nil {
					ctx.Logger().Warn("[msg: get auth cache fail] [error: db error] [err: %v]", err)
					return errors.ErrDBError
				}

				// todo: 验证auth.id 和 claims.Subject是否相同
				if !ok {
					ctx.Logger().Warn("[msg: get auth cache fail] [code: %s] [err: %v]", code, err)
					return nil
				}
				ctx.Logger().Info("[msg: auth info] [auth: %+v]", auth)
				// WithValue 设置uid，可使用ctx.GetUID()取出
				ctx.WithValue("uid", auth.UID)
				ctx.WithValue("appid", auth.Appid)
				return nil
			}
			validator := []jwt.Validator{
				jwt.VerifyIssuer("lazygo.dev"),
				jwt.VerifyAudience("m"),
				func(claims *jwt.Claims) error { return checkCode(claims.ID) },
			}
			_, _, err := jwt.ParseToken(utils.JwtDec, token, validator...)
			if err != nil {
				ctx.Logger().Warn("[msg: parse user token fail, try direct check code] [token: %s]", token)
			}
		}

		return next(ctx)
	})
}
