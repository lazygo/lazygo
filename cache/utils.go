package cache

import (
	"strconv"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

// 转换为string
func toString(value interface{}) string {
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
	case float32:
		str = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case float64:
		str = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	default:
		str = ""
	}
	return str
}
