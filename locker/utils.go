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

// randRange 返回min-max之间的随机数
func randRange(min uint64, max uint64) uint64 {
	if min > max {
		max, min = min, max
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + uint64(r.Int63n(int64(max-min)))
}
