package httpdns

import (
	"context"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestBaidu(t *testing.T) {
	assert := testify.New(t)
	b, err := baidu(map[string]string{"account": "186529", "secret": "kasCXQzsJzjZnsQm3N7v"})
	assert.Equal(err, nil)
	ipp, err := b.LookupIPAddr(context.Background(), "sh1.lazygo.dev")
	assert.Equal(err, nil)
	for _, item := range ipp {
		assert.Equal(item.Is4(), true)
		assert.Equal(item.String(), "150.158.45.167")
	}

}
