package model

import (
	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/model"
	goredis "github.com/redis/go-redis/v9"
)

type StoreCache struct {
	model.RedisModel
	ctx framework.Context
	key string
}

func NewStoreCache(ctx framework.Context, key string) *StoreCache {
	storeCache := &StoreCache{
		ctx: ctx,
		key: key,
	}
	storeCache.SetClient("lazygo-cache")
	return storeCache
}

func (mdl *StoreCache) Set(key string, value string) error {
	reply := mdl.RedisModel.HSet(mdl.ctx, mdl.key, key, value)
	if reply.Err() != nil {
		return reply.Err()
	}
	return nil
}

func (mdl *StoreCache) Get(field string) (string, error) {
	reply := mdl.RedisModel.HGet(mdl.ctx, mdl.key, field)
	if reply.Err() != nil {
		if reply.Err() == goredis.Nil {
			return "", nil
		}
		return "", reply.Err()
	}
	return reply.Val(), nil
}

func (mdl *StoreCache) Delete(fields ...string) error {
	reply := mdl.RedisModel.HDel(mdl.ctx, mdl.key, fields...)
	if reply.Err() != nil {
		return reply.Err()
	}
	return nil
}
