package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"

	"github.com/lazygo/lazygo/examples/framework"
	dbModel "github.com/lazygo/lazygo/examples/model/db"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
)

// CheckSign 校验签名
func CheckSign(fields ...string) server.MiddlewareFunc {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return framework.BaseHandlerFunc(func(ctx framework.Context) error {
			// 从请求中获取token,解析uid,如果没有传递token,uid设置为0
			sign := ctx.QueryParam("sign")
			appid := ctx.GetInt("appid")
			secret, ok := dbModel.AppidSecret[appid]
			if sign == "" || appid == 0 || !ok {
				return errors.ErrInvalidParams
			}

			data := make([]string, 0, len(fields))
			for _, item := range fields {
				data = append(data, fmt.Sprintf("%s=%s", item, ctx.QueryParam(item)))
			}
			slices.Sort(data)

			signStr := strings.Join(data, "&")
			wantSign := hex.EncodeToString(hmac.New(sha256.New, []byte(secret)).Sum([]byte(signStr)))
			if wantSign != sign {
				ctx.Logger().Warn("[msg: check sign fail] [want: %s] [has: %s] [signstr: %s]", wantSign, sign, signStr)
				return errors.ErrInvalidParams
			}

			return next(ctx)
		})
	}
}
