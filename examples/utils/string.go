package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/lazygo/pkg/feistel"
	"golang.org/x/exp/constraints"
)

func CutRune(str string, n int) string {
	r := []rune(str)
	if len(r) > n {
		r = r[:n]
	}
	return string(r)
}

const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const lowers = "0123456789abcdefghijklmnopqrstuvwxyz"
const base58Alphabets = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
const base33Alphabets = "123456789ABCDEFGHJKLMNPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

func RandStr(n int) string {
	return randStr(letters, n)
}

func RandLowerStr(n int) string {
	return randStr(lowers, n)
}

func randStr(l string, n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(l) {
			b[i] = l[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	builder := strings.Builder{}
	builder.Write(b)
	return builder.String()
}

func RandomBase33String(length int) string {
	return randStr(base33Alphabets, length)
}

func GenerateSerialNumber(mix uint64) string {
	// 时间戳+进程号+id+随机数
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(time.Now().UnixNano()))
	hash := md5.Sum(buf)
	str := fmt.Sprintf(
		"%06.6s%02.2s%04.4s%04.4s",
		strconv.FormatUint(binary.BigEndian.Uint64(hash[:]), 36),
		strconv.FormatUint(uint64(os.Getpid()%1296), 36),
		strconv.FormatUint(mix%1679616, 36),
		randStr(lowers, 4),
	)
	return strings.ToUpper(str)
}

func EncryptUID(u uint64) string {
	// 加密uid为字符串
	encrypted := feistel.EncryptUint32(uint32(u), 0x8ff0, 0x4435)
	return strconv.FormatUint(uint64(encrypted), 36)
}

func DecryptUID(s string) (uint64, error) {
	encrypted, err := strconv.ParseUint(s, 36, 64)
	if err != nil {
		return 0, err
	}
	return uint64(feistel.DecryptUint32(uint32(encrypted), 0x8ff0, 0x4435)), nil
}

func EncodeUint64(u uint64) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, u)
	v := Base58Encode(b)
	return string(v)
}

func DecodeUint64(v string) uint64 {
	b := make([]byte, 8)
	bu := Base58Decode([]byte(v))
	copy(b[8-len(bu):], bu)
	return binary.BigEndian.Uint64(b)
}

var DBC2SBC = unicode.SpecialCase{
	unicode.CaseRange{
		Lo: 0x3002, // Lo 全角句号
		Hi: 0x3002, // Hi 全角句号
		Delta: [unicode.MaxCase]rune{
			0,               // UpperCase
			0x002e - 0x3002, // LowerCase 转成半角句号
			0,               // TitleCase
		},
	},
	//
	unicode.CaseRange{
		Lo: 0xFF01, // 从全角！
		Hi: 0xFF19, // 到全角 9
		Delta: [unicode.MaxCase]rune{
			0,               // UpperCase
			0x0021 - 0xFF01, // LowerCase 转成半角
			0,               // TitleCase
		},
	},
}

var (
	EmailRegex  = regexp.MustCompile(`^([a-zA-Z0-9_.-])+@(([a-zA-Z0-9-])+\.)+([a-zA-Z0-9]{2,4})+$`)
	MobileRegex = regexp.MustCompile(`^1[345789]\d{9}$`)
	DigitRegex  = regexp.MustCompile(`^\d{1,20}$`)
	LetterRegex = regexp.MustCompile(`^[a-zA-Z]\w+$`)
	MD5Regex    = regexp.MustCompile(`^[0-9a-fA-F]{32}$`)
)

const (
	TypeUnknown = iota
	TypeEmail
	TypeMobile
	TypeId
	TypeName
)

func UsernameType(username string) int {
	if EmailRegex.MatchString(username) {
		return TypeEmail
	}
	if MobileRegex.MatchString(username) {
		return TypeMobile
	}
	if DigitRegex.MatchString(username) {
		return TypeId
	}
	if LetterRegex.MatchString(username) {
		return TypeName
	}
	return TypeUnknown
}

var (
	ErrInvalidAddr = errors.New("invalid addr")
)

func SplitHostPort(addr string) (string, int, error) {
	if !strings.Contains(addr, ":") {
		return addr, 0, nil
	}
	host, strport, err := net.SplitHostPort(addr)
	if err != nil {

		return "", 0, errors.Join(ErrInvalidAddr, err)
	}
	port, err := strconv.Atoi(strport)
	if err != nil {
		return "", 0, errors.Join(ErrInvalidAddr, err)
	}
	return host, port, nil
}

func SplitUnsigned[T constraints.Unsigned](str string) []T {
	var ids []T
	for item := range strings.SplitSeq(str, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		id, err := strconv.ParseUint(item, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, T(id))
	}
	return ids
}

func Split(str string) []string {
	var list []string
	for item := range strings.SplitSeq(str, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		list = append(list, item)
	}
	return list
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
		charIndex := strings.IndexByte(base58Alphabets, b)
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

func Mask(str string, head int, tail int) string {
	if head < 0 || tail < 0 {
		return str
	}
	l := len(str)
	if l <= head+tail {
		return str
	}
	var builder strings.Builder
	builder.Write([]byte(str[:head]))
	builder.Write(bytes.Repeat([]byte{'*'}, min(l-tail-head, 6)))
	builder.Write([]byte(str[l-tail:]))

	return builder.String()
}

func ContainsAny(str string, in []string) bool {
	return slices.ContainsFunc(in, func(v string) bool {
		return strings.Contains(str, v)
	})
}
