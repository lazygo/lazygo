package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ArtisanCloud/PowerLibs/v3/cache"
	"github.com/lazygo/lazygo/examples/model"
	goredis "github.com/redis/go-redis/v9"
)

// WechatCache 基于 Redis 的微信公众号缓存

type WechatCache struct {
	model.RedisModel
	ttl    int64
	format string
}

// NewWechatCache 创建公众号缓存实例
func NewWechatCache() *WechatCache {
	mdl := &WechatCache{
		ttl:    60,
		format: "wechat:%s",
	}
	mdl.SetClient("lazygo-cache")
	return mdl
}

// wrapKey 为 key 添加统一的前缀
func (mdl *WechatCache) wrapKey(key string) string {
	return fmt.Sprintf(mdl.format, key)
}

// ttlOrDefault ttl 小于等于 0 时返回默认过期时间
func (mdl *WechatCache) ttlOrDefault(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return time.Duration(mdl.ttl) * time.Second
	}
	return ttl
}

// Get 获取缓存，未命中时返回 cache.ErrCacheMiss
func (mdl *WechatCache) Get(key string, defaultValue any) (any, error) {
	b, err := mdl.RedisModel.Get(context.Background(), mdl.wrapKey(key)).Bytes()
	if err == goredis.Nil {
		return nil, cache.ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &defaultValue); err != nil {
		return nil, err
	}
	return defaultValue, nil
}

// Set 写入缓存，value 序列化为 JSON 后存储，expires 为过期时间
func (mdl *WechatCache) Set(key string, value any, expires time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return mdl.RedisModel.Set(context.Background(), mdl.wrapKey(key), b, expires).Err()
}

// Has 判断 key 是否存在
func (mdl *WechatCache) Has(key string) bool {
	n, err := mdl.Exists(context.Background(), mdl.wrapKey(key)).Result()
	return err == nil && n > 0
}

// AddNX 仅在 key 不存在时写入，写入成功返回 true
func (mdl *WechatCache) AddNX(key string, value any, ttl time.Duration) bool {
	b, err := json.Marshal(value)
	if err != nil {
		return false
	}
	ok, err := mdl.SetNX(context.Background(), mdl.wrapKey(key), b, mdl.ttlOrDefault(ttl)).Result()
	return err == nil && ok
}

// Add 仅在 key 不存在时写入，key 已存在时返回 error
func (mdl *WechatCache) Add(key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	ok, err := mdl.SetNX(context.Background(), mdl.wrapKey(key), b, mdl.ttlOrDefault(ttl)).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("this value has been actually added to the cache")
	}
	return nil
}

// Remember 获取缓存，未命中时执行 callback 并将结果写入缓存后返回
func (mdl *WechatCache) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	value, err := mdl.Get(key, nil)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return nil, err
	}
	if value != nil {
		return value, nil
	}

	value, err = callback()
	if err != nil {
		return nil, err
	}
	if err = mdl.Set(key, value, mdl.ttlOrDefault(ttl)); err != nil {
		return nil, err
	}
	return value, nil
}
