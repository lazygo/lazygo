package middleware

import (
	"github.com/lazygo/lazygo/server"
	"github.com/lazygo/pkg/trustip"
)

var ipExtractor = trustip.ExtractIPFromXFFHeader(
	trustip.XffsHeader(server.HeaderXForwardedFor),
	trustip.TrustIPRange("100.64.0.0/10"),
)

// TrustProxies
func TrustProxies(next server.HandlerFunc) server.HandlerFunc {
	return func(ctx server.Context) error {
		ip := ctx.Request().Header.Get("Eo-Client-Ip")
		if ip != "" {
			ctx.WithValue(server.HeaderXRealIP, ip)
		} else {
			ctx.WithValue(server.HeaderXRealIP, ipExtractor(ctx.Request()))
		}
		return next(ctx)
	}
}
