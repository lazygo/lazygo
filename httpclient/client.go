package httpclient

// http client 编译后代码包增加1.94MB

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/lzw"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/lazygo/lazygo/server"
)

const (
	HeaderRetayTimes  = "X-Control-Request-Retry-Times"
	HeaderTimeoutSec  = "X-Control-Request-Timeout-Sec"
	HeaderSpecifiedIP = "X-Control-Request-Specified-IP"
)

type Client struct {
	http.Client
}

func (hc *Client) ReadResponse(resp *http.Response) ([]byte, error) {

	var body io.ReadCloser = resp.Body

	switch resp.Header.Get(server.HeaderContentEncoding) {
	case "gzip":
		var err error
		body, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
	case "deflate":
		body = flate.NewReader(resp.Body)
	case "br":
		body = io.NopCloser(brotli.NewReader(resp.Body))
	case "zstd":
		var err error
		r, err := zstd.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		body = r.IOReadCloser()
	case "compress":
		body = lzw.NewReader(body, lzw.LSB, 8)
	default:
	}

	defer body.Close()

	return io.ReadAll(body)
}

func (hc *Client) ReadRequest(data []byte, encoding string) (io.Reader, error) {
	var buf bytes.Buffer
	switch encoding {
	case "gzip":
		w := gzip.NewWriter(&buf)
		defer w.Close()
		_, err := w.Write(data)
		return &buf, err
	case "deflate":
		w, err := flate.NewWriter(&buf, flate.DefaultCompression)
		if err != nil {
			return nil, err
		}
		defer w.Close()
		_, err = w.Write(data)
		return &buf, err
	case "br":
		w := brotli.NewWriter(&buf)
		defer w.Close()
		_, err := w.Write(data)
		return &buf, err
	case "zstd":
		w, err := zstd.NewWriter(&buf)
		if err != nil {
			return nil, err
		}
		defer w.Close()
		_, err = w.Write(data)
		return &buf, err
	case "compress":
		w := lzw.NewWriter(&buf, lzw.LSB, 8)
		defer w.Close()
		_, err := w.Write(data)
		return &buf, err
	default:
	}
	return bytes.NewReader(data), nil

}

func (hc *Client) Request(ctx context.Context, httpMethod string, url string, body []byte, headers map[string]string) ([]byte, int, error) {
	retryTimes := 3
	if headers[HeaderRetayTimes] != "" {
		retry, err := strconv.Atoi(headers[HeaderRetayTimes])
		if err != nil {
			return nil, 0, fmt.Errorf("%s=%s error: %w", HeaderRetayTimes, headers[HeaderRetayTimes], err)
		}
		retryTimes = retry
		if retryTimes < 0 {
			retryTimes = 0
		}
		if retryTimes > 10 {
			retryTimes = 10
		}
		delete(headers, HeaderRetayTimes)
	}
	if headers[HeaderTimeoutSec] != "" {
		sec, err := strconv.Atoi(headers[HeaderTimeoutSec])
		if err != nil {
			return nil, 0, fmt.Errorf("%s=%s error: %w", HeaderTimeoutSec, headers[HeaderTimeoutSec], err)
		}
		timeout := time.Duration(sec) * time.Second
		delete(headers, HeaderTimeoutSec)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if headers[HeaderSpecifiedIP] != "" {
		var ips []netip.Addr
		iplist := strings.Split(headers[HeaderSpecifiedIP], ",")
		for _, item := range iplist {
			ip, err := netip.ParseAddr(item)
			if err != nil {
				LogError("parse header %s=%s fail, %v", HeaderSpecifiedIP, headers[HeaderSpecifiedIP], err)
				continue
			}
			ips = append(ips, ip)
		}
		if len(ips) > 0 {
			ctx = context.WithValue(ctx, specifiedIPCtxKey{}, ips)
		}
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers[server.HeaderAcceptEncoding] = "gzip, deflate, br, zstd"

	var resp *http.Response

	// 注意此处不要使用 path.Join，因为BashUrl中可能带有http://，使用path.Join 会导致//合并为/
	// url := strings.TrimRight(hc.baseURL, "/") + "/" + strings.TrimLeft(uri, "/")
	for i := 0; i <= retryTimes; i++ {
		reqBody, err := hc.ReadRequest(body, headers[server.HeaderContentEncoding])
		if err != nil {
			return nil, 0, err
		}
		req, err := http.NewRequestWithContext(ctx, httpMethod, url, reqBody)
		if err != nil {
			return nil, 0, err
		}
		// request header
		for k, v := range headers {
			if k == server.HeaderContentLength {
				req.ContentLength, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, 0, fmt.Errorf("Content-Length error: %w", err)
				}
				continue
			}
			if k == "Host" {
				req.Host = k
				continue
			}
			req.Header.Add(k, v)
		}
		resp, err = hc.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return nil, 0, err
		}
	}

	defer resp.Body.Close()
	respBody, err := hc.ReadResponse(resp)
	return respBody, resp.StatusCode, err
}
