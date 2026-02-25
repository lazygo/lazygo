package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/lazygo/lazygo/httpclient/httpdns"
	testify "github.com/stretchr/testify/assert"
)

type BaseResp struct {
	Code  int    `json:"code"`
	Errno int    `json:"errno"`
	Msg   string `json:"msg"`
	Rid   uint64 `json:"rid"`
}

func TestBaidu(t *testing.T) {
	assert := testify.New(t)
	c := newhttpdns()
	body, status, err := c.Request(context.Background(), http.MethodGet, "https://sh1.lazygo.dev", nil, nil)
	assert.Equal(err, nil)
	assert.Equal(status, http.StatusNotFound)

	var resp BaseResp
	err = json.Unmarshal(body, &resp)
	assert.Equal(err, nil)
	assert.Equal(resp.Code, 404)
	assert.Equal(resp.Msg, "Not Found")

	_, status, err = c.Request(context.Background(), http.MethodGet, "https://www.lazygo.dev", nil, nil)
	assert.Equal(err, nil)
	assert.Equal(status, http.StatusOK)

}

func newhttpdns() *Client {

	httpdns.Init([]httpdns.Config{
		{
			Name:    "baidu",
			Adapter: "baidu",
			Option:  map[string]string{"account": "186529", "secret": "kasCXQzsJzjZnsQm3N7v"},
		},
	}, "baidu")

	conf := &Config{
		HTTPDNSAdapter: "baidu",
	}
	return New(conf).Client(&HttpConfig{Timeout: 30 * time.Second})
}
