package worker

import (
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/httpclient"
)

func Bootstrap(ctx framework.Context) {
	httpclient.LogDebug = ctx.Logger().Debug
	httpclient.LogError = ctx.Logger().Error

	go func() {
		// TODO: 添加worker初始化
	}()
}
