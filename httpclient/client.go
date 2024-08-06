package httpclient

// http client 编译后代码包增加1.94MB

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lazygo/lazygo/server"
)

const (
	RetayTimes = "X-Control-Request-Retry-Times"
	TimeoutSec = "X-Control-Request-Timeout-Sec"
)

type Client struct {
	http.Client
	baseURL string
}

func (hc *Client) Request(ctx context.Context, httpMethod string, uri string, body []byte, headers map[string]string) ([]byte, int, error) {

	retryTimes := 3
	if headers[RetayTimes] != "" {
		retry, err := strconv.Atoi(headers[RetayTimes])
		if err != nil {
			return nil, 0, fmt.Errorf("%s error: %w", RetayTimes, err)
		}
		retryTimes = retry
		delete(headers, RetayTimes)
	}
	if headers[TimeoutSec] != "" {
		sec, err := strconv.Atoi(headers[TimeoutSec])
		if err != nil {
			return nil, 0, fmt.Errorf("%s error: %w", TimeoutSec, err)
		}
		timeout := time.Duration(sec) * time.Second
		delete(headers, TimeoutSec)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var resp *http.Response

	// 注意此处不要使用 path.Join，因为BashUrl中可能带有http://，使用path.Join 会导致//合并为/
	url := strings.TrimRight(hc.baseURL, "/") + "/" + strings.TrimLeft(uri, "/")
	for i := 1; i <= retryTimes; i++ {
		req, err := http.NewRequestWithContext(ctx, httpMethod, url, bytes.NewReader(body))
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

	respBody, err := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, err
}
