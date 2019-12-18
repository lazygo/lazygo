package mysql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func DefaultString(str string, def string) string {
	if str == "" {
		return def
	}
	return str
}

func DefaultInt64(i int64, def int64) int64 {
	if i == 0 {
		return def
	}
	return i
}

func ContainInArray(str string, arr []string) bool {
	for _, s := range arr {
		if strings.Index(s, str) != -1 {
			return true
		}
	}
	return false
}

func Addslashes(v string) string {
	pos := 0
	buf := make([]byte, len(v)*2)
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c == '\'' || c == '"' || c == '\\' {
			buf[pos] = '\\'
			buf[pos+1] = c
			pos += 2
		} else {
			buf[pos] = c
			pos++
		}
	}
	return string(buf[:pos])
}

func ToString(value interface{}) string {
	var str string
	switch value.(type) {
	case string:
		str = value.(string)
	case []byte:
		str = string(value.([]byte))
	case int:
		str = strconv.Itoa(value.(int))
	case int64:
		str = strconv.FormatInt(value.(int64), 10)
	default:
		str = ""
	}
	return str
}

func ToInt64(value interface{}) int64 {
	val := reflect.ValueOf(value)
	var d int64
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = val.Int()
	case uint, uint8, uint16, uint32, uint64:
		d = int64(val.Uint())
	case string:
		d, err = strconv.ParseInt(val.String(), 10, 64)
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	return d
}
