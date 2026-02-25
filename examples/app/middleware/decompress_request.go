package middleware

import (
	"compress/gzip"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
)

// DEBUG
func DecompressRequest(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		r := ctx.Request()
		// 检查请求头是否为 gzip 压缩
		if r.Header.Get("Content-Encoding") == "gzip" {
			// 创建 gzip.Reader 解压请求体
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				return errors.ErrInternalServerError
			}
			defer gz.Close()
			// 替换请求体为解压后的流
			r.Body = gz
		}
		return next(ctx)
	})
}
