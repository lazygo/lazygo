package cache

import (
	"fmt"

	"github.com/lazygo/lazygo/examples/model"
	"github.com/lazygo/lazygo/examples/utils"
)

type AuthUserCache struct {
	model.CacheModel
	ttl int64
}

type AuthUserData struct {
	UID int64 `json:"uid"`
}

func NewAuthUserCache() *AuthUserCache {
	cacheAuthUser := &AuthUserCache{
		ttl: 3600 * 24 * 7,
	}
	cacheAuthUser.SetCache("hdrds")
	return cacheAuthUser
}

func (mdl *AuthUserCache) Get(token string) (*AuthUserData, bool, error) {
	key := fmt.Sprintf(utils.CACHE_AUTH_TOKEN, token)

	var info AuthUserData
	exists, err := mdl.Cache.Get(key, &info)
	if err != nil {
		return nil, false, err
	}
	return &info, exists, nil
}

func (mdl *AuthUserCache) Forget(token string) error {
	key := fmt.Sprintf(utils.CACHE_AUTH_TOKEN, token)
	return mdl.Cache.Forget(key)
}

func (mdl *AuthUserCache) Set(uid int64) (string, error) {
	token := utils.RandStr(32)
	key := fmt.Sprintf(utils.CACHE_AUTH_TOKEN, token)

	info := AuthUserData{
		UID: uid,
	}
	err := mdl.Cache.Set(key, info, mdl.ttl)
	if err != nil {
		return "", err
	}
	return token, nil
}
