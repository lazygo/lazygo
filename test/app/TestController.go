package app

import (
	"github.com/lazygo/lazygo/engine"
)

type TestController struct{}

func (t TestController) TestResponseAction(ctx engine.Context) error {
	return ctx.JSON(401, map[string]interface{}{
		"aa": "bba",
	})
}
