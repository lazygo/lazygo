package locker

import (
	"math/rand"
	"time"
)

// randomToken 生成随机token
func randomToken() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Uint64()
}
