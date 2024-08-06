package httpdns

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"time"
)

type BaiduHTTPDNS struct {
	Endpoint string `json:"endpoint"`
	Account  string `json:"account"`
	Secret   string `json:"secret"`
}

type BaiduDNSData struct {
	IP  []string `json:"ip"`
	TTL int64    `json:"ttl"`
}

type BaiduHTTPDNSResp struct {
	Clientip  string                  `json:"clientip"`
	Data      map[string]BaiduDNSData `json:"data"`
	Msg       string                  `json:"msg"`
	Timestamp int64                   `json:"timestamp"`
}

var client = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Resolver: &net.Resolver{},
			}
			return dialer.DialContext(ctx, network, addr)
		},
	},
}

func baidu(args map[string]string) (HTTPDNS, error) {
	account, ok1 := args["account"]
	secret, ok2 := args["secret"]
	if !ok1 || !ok2 {
		return nil, ErrInvalidHTTPDNSAdapterParams
	}
	return &BaiduHTTPDNS{
		Endpoint: "",
		Account:  account,
		Secret:   secret,
	}, nil
}

func (httpdns *BaiduHTTPDNS) LookupIPAddr(ctx context.Context, host string) ([]netip.Addr, error) {
	params := &url.Values{}
	params.Set("account_id", httpdns.Account)
	params.Set("dn", host)
	timestamp := time.Now().Unix()
	params.Set("t", strconv.FormatInt(timestamp, 10))
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%d", host, httpdns.Secret, timestamp))))
	params.Set("sign", sign)
	url := httpdns.Endpoint + "?" + params.Encode()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("request http dns fail, uri: %s, err: %w", url, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request http dns fail, uri: %s, err: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request http dns fail, uri: %s, err: %w", url, err)
	}
	var data BaiduHTTPDNSResp
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("request http dns fail, uri: %s, err: %w", url, err)
	}
	if data.Msg != "ok" {
		return nil, fmt.Errorf("request http dns fail, uri: %s, err msg: %s", url, data.Msg)
	}
	ipList := data.Data[host].IP
	if len(ipList) == 0 {
		return nil, fmt.Errorf("request http dns ip list empty, uri: %s", url)
	}

	var ips []netip.Addr
	for _, addr := range ipList {
		ip, err := netip.ParseAddr(addr)
		if err != nil || !ip.IsValid() {
			continue
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func init() {
	// 注册适配器
	registry.Add("baidu", baidu)
}
