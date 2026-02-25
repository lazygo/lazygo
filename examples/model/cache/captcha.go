package model

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/examples/model"
)

var (
	SmsTpl = map[string]string{
		"register": "2295547",
		"bind":     "2295549",
		"forget":   "2363250",
	}
)

type CaptchaCache struct {
	model.CacheModel
	ttl    int64
	format string
}

func NewCaptchaCache(ctx framework.Context) *CaptchaCache {
	cacheCaptcha := &CaptchaCache{
		ttl:    1800,
		format: "user:captcha:%s:%s",
	}
	cacheCaptcha.SetCache("lazygo-cache")
	return cacheCaptcha
}

func (mdl *CaptchaCache) Store(uk string, opname string) (string, error) {
	code := strconv.Itoa(rand.Intn(900000) + 100000)
	key := fmt.Sprintf(mdl.format, uk, opname)
	err := mdl.Cache.Set(key, code, mdl.ttl)
	return code, err
}

func (mdl *CaptchaCache) Load(uk string, opname string) (string, error) {
	key := fmt.Sprintf(mdl.format, uk, opname)
	var code string
	_, err := mdl.Cache.Get(key, &code)
	if err != nil {
		return "", err
	}
	return code, err
}
