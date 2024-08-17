package httpclient

// http client 编译后代码包增加1.94MB

import (
	"cmp"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/lazygo/lazygo/httpclient/httpdns"
	"github.com/patrickmn/go-cache"
	"go4.org/netipx"
)

var (
	dnsResolverProto     = "udp" // Protocol to use for the DNS resolver
	dnsResolverTimeoutMs = 5000  // Timeout (ms) for the DNS resolver (optional)
)

var dnscache = cache.New(2*time.Minute, 10*time.Minute)

type Manager struct {
	resolver *net.Resolver
	httpdns  httpdns.HTTPDNS
}

var (
	LogDebug = func(format string, v ...any) { log.Printf(format, v...) }
	LogError = func(format string, v ...any) { log.Printf(format, v...) }
)

type specifiedIPCtxKey struct{}
type SpecifiedIP []netip.Addr

type Config struct {
	DNSResolverAddr string `json:"dns"`
	HTTPDNSAdapter  string `json:"httpdns"`
}

func New(conf *Config) *Manager {
	m := &Manager{}
	if conf.DNSResolverAddr != "" {
		m.resolver = resolver(conf.DNSResolverAddr)
	}
	if conf.HTTPDNSAdapter != "" {
		httpdns, err := httpdns.Instance(conf.HTTPDNSAdapter)
		if err != nil {
			m.httpdns = httpdns
		}
	}
	return m
}

func (m *Manager) Transport(timeout time.Duration) *http.Transport {
	connTimeout := 5 * time.Second
	dialer := &net.Dialer{
		Timeout:  connTimeout,
		Resolver: m.resolver,
	}

	return &http.Transport{
		MaxIdleConnsPerHost: -1,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if ipp, err := netip.ParseAddrPort(addr); err != nil || !ipp.IsValid() {
				separator := strings.LastIndex(addr, ":")
				if ips, ok := ctx.Value(specifiedIPCtxKey{}).(SpecifiedIP); ok && len(ips) > 0 {
					// specified ip from request
					addr = ips[rand.Intn(len(ips))].String() + addr[separator:]
				} else {
					ips, err := m.lookupHost(ctx, addr[:separator])
					if err != nil {
						LogError("resolve dns fail: %v", err)
						return nil, fmt.Errorf("resolve dns fail: %w", err)
					}
					addr = ips[rand.Intn(len(ips))].String() + addr[separator:]
				}
			}

			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			err = conn.SetDeadline(time.Now().Add(timeout))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
}

func (m *Manager) lookupHost(ctx context.Context, host string) ([]netip.Addr, error) {

	var addrs []netip.Addr

	if ips, found := dnscache.Get(host); found {
		LogDebug("lookup host found in cache: %s", ips.([]netip.Addr))
		return ips.([]netip.Addr), nil
	}

	//获取dns IP
	ips, err := cmp.Or(m.resolver, net.DefaultResolver).LookupIPAddr(context.Background(), host)
	if err != nil {
		LogError("lookup host %s fail, try httpdns: %v", host, err)
	}

	if err == nil && len(ips) > 0 {
		for _, ip := range ips {
			LogDebug("lookupHost: %s", ip.String())
			addr, ok := netipx.FromStdIP(ip.IP)
			if !ok {
				LogError("lookupHost: %s", ip.String())
				continue
			}
			addrs = append(addrs, addr)

		}
	}
	if len(addrs) > 0 {
		dnscache.Set(host, addrs, cache.DefaultExpiration)
		return addrs, nil
	}
	LogDebug("resolve dns fail, try httpdns")
	// 请求其他的DNS服务
	if m.httpdns != nil {
		addrs, err = m.httpdns.LookupIPAddr(ctx, host)
		if err == nil && len(addrs) > 0 {
			dnscache.Set(host, addrs, cache.DefaultExpiration)
			return addrs, nil
		}
	}

	return nil, fmt.Errorf("no record found for %s", host)
}

func resolver(dns string) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
			}
			return d.DialContext(ctx, dnsResolverProto, dns)
		},
	}
}

func (m *Manager) Client(timeout time.Duration) *Client {
	client := http.Client{Transport: m.Transport(timeout)}
	client.Timeout = timeout
	return &Client{
		Client: client,
	}
}
