package locker

import (
	"crypto/rand"
	"encoding/binary"
)

// randomToken 生成随机token
func randomToken() uint64 {
	var b [8]byte
	// crypto/rand保证跨进程不碰撞，避免误删他人锁
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint64(b[:])
}
