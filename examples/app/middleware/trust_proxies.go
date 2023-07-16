package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/server"
)

type ipChecker struct {
	trustLoopback    bool
	trustLinkLocal   bool
	trustPrivateNet  bool
	trustExtraRanges []*net.IPNet
}

// TrustOption is config for which IP address to trust
type TrustOption func(*ipChecker)

// TrustLoopback configures if you trust loopback address (default: true).
func TrustLoopback(v bool) TrustOption {
	return func(c *ipChecker) {
		c.trustLoopback = v
	}
}

// TrustLinkLocal configures if you trust link-local address (default: true).
func TrustLinkLocal(v bool) TrustOption {
	return func(c *ipChecker) {
		c.trustLinkLocal = v
	}
}

// TrustPrivateNet configures if you trust private network address (default: true).
func TrustPrivateNet(v bool) TrustOption {
	return func(c *ipChecker) {
		c.trustPrivateNet = v
	}
}

// TrustIPRange add trustable IP ranges using CIDR notation.
func TrustIPRange(ipRange string) TrustOption {
	_, ipNet, err := net.ParseCIDR(ipRange)
	return func(c *ipChecker) {
		if err != nil {
			framework.App().Logger.Panicf("[msg: add trust ip range fail] [ipRange: %s]", ipRange)
			return
		}
		c.trustExtraRanges = append(c.trustExtraRanges, ipNet)
	}
}

func newIPChecker(configs []TrustOption) *ipChecker {
	checker := &ipChecker{trustLoopback: true, trustLinkLocal: true, trustPrivateNet: true}
	for _, configure := range configs {
		configure(checker)
	}
	return checker
}

func (c *ipChecker) trust(ip net.IP) bool {
	if c.trustLoopback && ip.IsLoopback() {
		return true
	}
	if c.trustLinkLocal && ip.IsLinkLocalUnicast() {
		return true
	}
	if c.trustPrivateNet && ip.IsPrivate() {
		return true
	}
	for _, trustedRange := range c.trustExtraRanges {
		if trustedRange.Contains(ip) {
			return true
		}
	}
	return false
}

type IPExtractor func(*http.Request) string

// ExtractIPDirect extracts IP address using actual IP address.
// Use this if your server faces to internet directory (i.e.: uses no proxy).
func ExtractIPDirect() IPExtractor {
	return extractIP
}

func extractIP(req *http.Request) string {
	ra, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ra
}

// ExtractIPFromRealIPHeader extracts IP address using x-real-ip header.
// Use this if you put proxy which uses this header.
func ExtractIPFromRealIPHeader(options ...TrustOption) IPExtractor {
	checker := newIPChecker(options)
	return func(req *http.Request) string {
		realIP := req.Header.Get(server.HeaderXRealIP)
		if realIP != "" {
			realIP = strings.TrimPrefix(realIP, "[")
			realIP = strings.TrimSuffix(realIP, "]")
			if ip := net.ParseIP(realIP); ip != nil && checker.trust(ip) {
				return realIP
			}
		}
		return extractIP(req)
	}
}

// ExtractIPFromXFFHeader extracts IP address using x-forwarded-for header.
// Use this if you put proxy which uses this header.
// This returns nearest untrustable IP. If all IPs are trustable, returns furthest one (i.e.: XFF[0]).
func ExtractIPFromXFFHeader(options ...TrustOption) IPExtractor {
	checker := newIPChecker(options)
	return func(req *http.Request) string {
		directIP := extractIP(req)
		xffs := req.Header[server.HeaderXForwardedFor]
		if len(xffs) == 0 {
			return directIP
		}
		ips := append(strings.Split(strings.Join(xffs, ","), ","), directIP)
		for i := len(ips) - 1; i >= 0; i-- {
			ips[i] = strings.TrimSpace(ips[i])
			ips[i] = strings.TrimPrefix(ips[i], "[")
			ips[i] = strings.TrimSuffix(ips[i], "]")
			ip := net.ParseIP(ips[i])
			if ip == nil {
				// Unable to parse IP; cannot trust entire records
				return directIP
			}
			if !checker.trust(ip) {
				return ip.String()
			}
		}
		// All of the IPs are trusted; return first element because it is furthest from server (best effort strategy).
		return strings.TrimSpace(ips[0])
	}
}

// TrustProxies
func TrustProxies(next server.HandlerFunc) server.HandlerFunc {
	ipExtractor := ExtractIPFromXFFHeader(
		TrustIPRange("100.64.0.0/10"),
	)

	return func(ctx server.Context) error {
		ctx.WithValue("real_ip", ipExtractor(ctx.Request()))
		return next(ctx)
	}
}
