package router

import (
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/utils/errors"
	"github.com/lazygo/lazygo/server"
)

var ErrTo500 = func(next server.HandlerFunc) server.HandlerFunc {
	return framework.BaseHandlerFunc(func(ctx framework.Context) error {
		err := next(ctx)
		if err != nil {
			return errors.ErrInternalServerError
		}
		return nil
	})
}

// ThirdRouter 第三方服务
func ThirdRouter(g *server.Group) {

	// 微信支付回调接收 /api/payment/wxpay_qrcode/notify
	// app.Any("/api/payment/:pay_way/notify", server.Controller(controller.PaymentController{}), ErrTo500)
	// app.Any("/api/wechat/notify", server.Controller(controller.WechatController{}), ErrTo500)

}
