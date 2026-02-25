package model

import (
	"context"
	"fmt"
	"time"

	"github.com/lazygo-dev/lazygo/examples/model"
)

type WechatCache struct {
	model.RedisModel
	ttl    int64
	format string
}

func NewWechatCache() *WechatCache {
	mdl := &WechatCache{
		ttl:    60,
		format: "wechat:%s",
	}
	mdl.SetClient("lazygo-cache")
	return mdl
}

func (mdl *WechatCache) Get(key string) interface{} {
	key = fmt.Sprintf(mdl.format, key)
	result, err := mdl.RedisModel.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}
	return result
}

func (mdl *WechatCache) Set(key string, val interface{}, timeout time.Duration) error {
	key = fmt.Sprintf(mdl.format, key)
	return mdl.RedisModel.Set(context.Background(), key, val, timeout).Err()
}

func (mdl *WechatCache) IsExist(key string) bool {
	key = fmt.Sprintf(mdl.format, key)
	result, err := mdl.Exists(context.Background(), key).Result()
	if err != nil {
		return false
	}
	return result > 0
}

func (mdl *WechatCache) Delete(key string) error {
	key = fmt.Sprintf(mdl.format, key)
	return mdl.Del(context.Background(), key).Err()
}
