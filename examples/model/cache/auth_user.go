package model

import (
	"fmt"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/model"
	"github.com/lazygo/lazygo/examples/utils"
	"github.com/lazygo/pkg/goutils"
)

type AuthUserCache struct {
	model.CacheModel
	ttl int64
}

type AuthUserData struct {
	UID   uint64 `json:"uid"`
	Appid int    `json:"appid"`
}

func NewAuthUserCache(ctx framework.Context) *AuthUserCache {
	cacheAuthUser := &AuthUserCache{
		ttl: 3600 * 24 * 365,
	}
	cacheAuthUser.SetCache("lazygo-cache")
	return cacheAuthUser
}

func (mdl *AuthUserCache) Get(token string) (*AuthUserData, bool, error) {
	key := fmt.Sprintf(utils.CacheAuthToken, token)

	var info AuthUserData
	exists, err := mdl.Cache.Get(key, &info)
	if err != nil {
		return nil, false, err
	}
	return &info, exists, nil
}

func (mdl *AuthUserCache) Forget(token string) error {
	key := fmt.Sprintf(utils.CacheAuthToken, token)
	return mdl.Cache.Forget(key)
}

func (mdl *AuthUserCache) Set(appid int, uid uint64) (string, error) {
	token := goutils.RandStr(32)
	key := fmt.Sprintf(utils.CacheAuthToken, token)

	info := AuthUserData{
		UID:   uid,
		Appid: appid,
	}
	err := mdl.Cache.Set(key, info, mdl.ttl)
	if err != nil {
		return "", err
	}
	return token, nil
}
