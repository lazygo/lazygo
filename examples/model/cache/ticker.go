package model

import (
	"context"
	"fmt"
	"time"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/model"
)

const (
	ThrottleDuration = "X-Throttle-Duration"
	ThrottleNum      = "X-Throttle-Num"
)

type TickerCache struct {
	model.RedisModel
	ttl    int64
	format string
}

func NewTickerCache(ctx framework.Context) *TickerCache {
	mdl := &TickerCache{
		ttl:    1800,
		format: "system:ticker:%s:%d",
	}
	mdl.SetClient("lazygo-cache")
	return mdl
}

func (mdl *TickerCache) Can(scene string, n int64, duration time.Duration) (bool, error) {
	window := time.Now().Unix() / int64(duration.Seconds())
	key := fmt.Sprintf(mdl.format, scene, window)
	reply := mdl.Incr(context.Background(), key)
	if reply.Err() != nil {
		return false, reply.Err()
	}
	mdl.Expire(context.Background(), key, duration)

	val := reply.Val()
	if val > n {
		return false, nil
	}

	return true, nil
}
