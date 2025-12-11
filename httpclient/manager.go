package httpclient

// http client 编译后代码包增加1.94MB

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
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

type HttpConfig struct {
	Timeout             time.Duration
	TLSConfig           *tls.Config
	ProxyURL            *url.URL
	DialerTimeout       time.Duration
	MaxIdleConnsPerHost int
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

func (m *Manager) Transport(config *HttpConfig) *http.Transport {
	dialer := &net.Dialer{
		Timeout:  config.DialerTimeout,
		Resolver: m.resolver,
	}
	if config.DialerTimeout == 0 {
		dialer.Timeout = 5 * time.Second
	}

	tlsClientConfig := &tls.Config{InsecureSkipVerify: true}
	if config.TLSConfig != nil {
		tlsClientConfig = config.TLSConfig
	}

	tr := &http.Transport{
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		TLSClientConfig:     tlsClientConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			separator := strings.LastIndex(addr, ":")
			if separator == -1 {
				return nil, fmt.Errorf("invalid address: %s", addr)
			}
			port, err := strconv.Atoi(addr[separator+1:])
			if err != nil {
				return nil, fmt.Errorf("parse port %s fail: %w", addr[separator+1:], err)
			}
			addrs := make([]netip.AddrPort, 0)
			if ipp, err := netip.ParseAddrPort(addr); err != nil || !ipp.IsValid() {
				if ips, ok := ctx.Value(specifiedIPCtxKey{}).(SpecifiedIP); ok && len(ips) > 0 {
					// specified ip from request
					// addr = ips[rand.Intn(len(ips))].String() + addr[separator:]
					for _, ip := range ips {
						addrs = append(addrs, netip.AddrPortFrom(ip, uint16(port)))
					}
				} else {
					ips, err := m.lookupHost(ctx, addr[:separator])
					if err != nil {
						LogError("resolve dns fail: %v", err)
						return nil, fmt.Errorf("resolve dns fail: %w", err)
					}
					for _, ip := range ips {
						addrs = append(addrs, netip.AddrPortFrom(ip, uint16(port)))
					}
				}
			} else {
				addrs = append(addrs, ipp)
			}

			if len(addrs) == 0 {
				return nil, fmt.Errorf("no record found for %s", addr)
			}

			var conn net.Conn
			for _, addr := range addrs {
				conn, err = dialer.DialContext(ctx, network, addr.String())
				if err != nil {
					continue
				}
				break
			}
			if err != nil {
				return conn, err
			}

			return conn, nil
		},
	}

	if config.ProxyURL != nil {
		if strings.HasPrefix(strings.ToLower(config.ProxyURL.Scheme), "env") {
			tr.Proxy = http.ProxyFromEnvironment
		} else {
			tr.Proxy = http.ProxyURL(config.ProxyURL)
		}
	}

	return tr
}

func (m *Manager) lookupHost(ctx context.Context, host string) ([]netip.Addr, error) {

	var addrs []netip.Addr

	if ips, found := dnscache.Get(host); found {
		LogDebug("lookup host %s found in cache: %s", host, ips.([]netip.Addr))
		return ips.([]netip.Addr), nil
	}

	//获取dns IP
	resolver := m.resolver
	if m.resolver == nil {
		resolver = net.DefaultResolver
	}
	ips, err := resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		LogError("lookup host %s fail, try httpdns: %v", host, err)
	}

	if err == nil && len(ips) > 0 {
		for _, ip := range ips {
			LogDebug("lookup host %s: %s", host, ip.String())
			addr, ok := netipx.FromStdIP(ip.IP)
			if !ok {
				LogError("lookup host %s fail: %s", host, ip.String())
				continue
			}
			addrs = append(addrs, addr)
		}
	}
	if len(addrs) > 0 {
		dnscache.Set(host, addrs, cache.DefaultExpiration)
		return addrs, nil
	}
	// 请求其他的DNS服务
	if m.httpdns != nil {
		LogDebug("resolve dns fail, try httpdns")
		addrs, err = m.httpdns.LookupIPAddr(ctx, host)
		if err != nil {
			LogError("resolve httpdns fail")
		}
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

func (m *Manager) Client(config *HttpConfig) *Client {
	client := http.Client{Transport: m.Transport(config)}
	client.Timeout = config.Timeout
	return &Client{
		Client: client,
	}
}
