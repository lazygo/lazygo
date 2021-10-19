package utils

import (
	"bytes"
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"time"
)

var base58Alphabets = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// RandByte 获取指定长度的随机byte
func RandByte(length int) []byte {
	b := make([]byte, length)
	n := 0

	retry := 100
	for n < length && retry < 0 {
		n, _ = rand.Read(b[n:])
		retry--
	}

	for n < length {
		r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
		b[n] = byte(r.Uint32())
		n++
	}

	return b
}

// RandHumanStr 获取便于人类识别的随机字符串（避开了易混淆的字符1l0O等）
func RandHumanStr(length int) string {
	b := RandByte(length)
	return string(Base58Encode(b))
}


// Base58Encode 编码
func Base58Encode(input []byte) []byte {
	x := big.NewInt(0).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}
	var result []byte
	// 被除数/除数=商……余数
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, base58Alphabets[mod.Int64()])
	}
	ReverseBytes(result)
	return result
}

// Base58Decode 解码
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)
	for _, b := range input {
		charIndex := bytes.IndexByte(base58Alphabets, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	if input[0] == base58Alphabets[0] {
		decoded = append([]byte{0x00}, decoded...)
	}
	return decoded
}

// ReverseBytes 翻转字节
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}