package utils

import (
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
)

func CutRune(str string, n int) string {
	r := []rune(str)
	if len(r) > n {
		r = r[:n]
	}
	return string(r)
}

func InSlice(haystack []string, needle string) bool {
	for _, e := range haystack {
		if e == needle {
			return true
		}
	}
	return false
}

const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

func RandStr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	builder := strings.Builder{}
	builder.Write(b)
	return builder.String()
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
