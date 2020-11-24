package utils

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"time"
)

// 获取指定长度的随机字符串
func RandId(length int) string {
	b := make([]byte, length)
	n, err := rand.Read(b)

	if n != length {
		err = fmt.Errorf("Only generated %d random bytes, %d requested", n, length)
		CheckError(err)
	}

	id := fmt.Sprintf("%x", b)
	return id
}

// 获取便于人类识别的随机字符串（避开了易混淆的字符1l0O等）
func RandStr(l int) string {
	str := "23456789abcdefghijkmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
