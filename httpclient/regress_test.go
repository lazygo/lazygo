package httpclient

// httpclient缺陷修复的回归测试

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/lazygo/lazygo/httpclient/httpdns"
	testify "github.com/stretchr/testify/assert"
)

// Host头修复：req.Host应取header的值而非键名
func TestRequestHostHeader(t *testing.T) {
	assert := testify.New(t)

	var gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&Config{}).Client(&HttpConfig{Timeout: 5 * time.Second})
	_, status, err := c.Request(context.Background(), http.MethodGet, srv.URL, nil, map[string]string{
		"Host": "custom.example.com",
	})
	assert.Nil(err)
	assert.Equal(http.StatusOK, status)
	assert.Equal("custom.example.com", gotHost)
}

// 控制头修复：X-Control-Request-Specified-IP不应转发给目标服务器
func TestSpecifiedIPHeaderNotForwarded(t *testing.T) {
	assert := testify.New(t)

	var leaked string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get(HeaderSpecifiedIP)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(&Config{}).Client(&HttpConfig{Timeout: 5 * time.Second})
	_, status, err := c.Request(context.Background(), http.MethodGet, srv.URL, nil, map[string]string{
		HeaderSpecifiedIP: "127.0.0.1",
	})
	assert.Nil(err)
	assert.Equal(http.StatusOK, status)
	assert.Equal("", leaked)
}

// httpdns启用修复：Instance成功时应赋值给Manager
func TestManagerHTTPDNSEnabled(t *testing.T) {
	assert := testify.New(t)

	err := httpdns.Init([]httpdns.Config{
		{
			Name:    "regress",
			Adapter: "baidu",
			Option:  map[string]string{"account": "dummy", "secret": "dummy"},
		},
	}, "regress")
	assert.Nil(err)

	m := New(&Config{HTTPDNSAdapter: "regress"})
	assert.NotNil(m.httpdns)
}

// 指定IP直连修复：ctx中的SpecifiedIP应被Transport正确识别（存储与断言类型一致）
func TestSpecifiedIPDial(t *testing.T) {
	assert := testify.New(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(err)
	defer ln.Close()

	lnAddr, err := netip.ParseAddrPort(ln.Addr().String())
	assert.Nil(err)

	m := &Manager{}
	tr := m.Transport(&HttpConfig{DialerTimeout: 2 * time.Second})

	// .invalid 保证不可解析，只有specified-IP路径能建立连接；端口保持目标地址的端口
	ctx := context.WithValue(context.Background(), specifiedIPCtxKey{}, SpecifiedIP{lnAddr.Addr()})
	conn, err := tr.DialContext(ctx, "tcp", fmt.Sprintf("unresolvable-host.invalid:%d", lnAddr.Port()))
	if !assert.Nil(err) {
		return
	}
	defer conn.Close()

	remote, err := netip.ParseAddrPort(conn.RemoteAddr().String())
	assert.Nil(err)
	assert.Equal(lnAddr.Port(), remote.Port())
}

var _ = fmt.Sprintf
