package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/httpclient"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrRequestApiFail = errors.New("request api fail")
	ErrRequestFail    = errors.New("request fail")
	ErrInvaildContent = errors.New("response content error")
)

type BaseResp struct {
	Code  int    `json:"code"`
	Errno int    `json:"errno"`
	Msg   string `json:"msg"`
	Rid   uint64 `json:"rid"`
	Time  int64  `json:"t"`
}
type TxModel[T any] struct {
	mysql.TxModel[T]
}
type CacheModel struct {
	cache.Cache
}

func (m *CacheModel) SetCache(name string) {
	m.Cache = m.cache(name)
}

func (m *CacheModel) cache(name string) cache.Cache {
	instance, err := cache.Instance(name)
	if err != nil {
		panic(fmt.Sprintln(name, err))
	}
	return instance
}

type RedisModel struct {
	*goredis.Client
}

func (m *RedisModel) SetClient(name string) {
	m.Client = m.client(name)
}

func (m *RedisModel) client(name string) *goredis.Client {
	instance, err := redis.Client(name)
	if err != nil {
		panic(fmt.Sprintln(name, err))
	}
	return instance
}

type RPCModel struct {
	BaseURL string
	IPs     []netip.Addr
}

var (
	client     *httpclient.Client
	clientOnce sync.Once
)

func DefaultClient() *httpclient.Client {
	clientOnce.Do(func() {
		client = httpclient.New(&httpclient.Config{
			// DNSResolverAddr:"",
			HTTPDNSAdapter: "baidu",
		}).Client(&httpclient.HttpConfig{
			Timeout:             10 * time.Second,
			MaxIdleConnsPerHost: 20,
		})
	})
	return client
}

func (mdl *RPCModel) Request(ctx framework.Context, method string, uri string, data []byte, v any) error {
	url := strings.TrimRight(mdl.BaseURL, "/") + "/" + strings.TrimLeft(uri, "/")

	headers := map[string]string{"Authorization": "sign"}

	var ips []string
	for _, ip := range mdl.IPs {
		ips = append(ips, ip.String())
	}
	if len(ips) > 0 {
		headers[httpclient.HeaderSpecifiedIP] = strings.Join(ips, ",")
	}

	ctx.Logger().Debug("rpc request: %s %s %v", method, url, headers)

	body, status, err := DefaultClient().Request(ctx, method, url, data, headers)
	ctx.Logger().Debug("rpc response: %s", string(body))
	if err != nil {
		return errors.Join(ErrRequestFail, err)
	}
	var resp struct {
		BaseResp
		Data json.RawMessage `json:"data"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return errors.Join(ErrInvaildContent, err)
	}
	if resp.Errno != 0 {
		return fmt.Errorf("%w: %s", ErrRequestApiFail, resp.Msg)
	}
	if status != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrRequestApiFail, "network error")
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(resp.Data, v)
}
