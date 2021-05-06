package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// 转换为字符串
// 如果value不能转换为字符串，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为空字符串""
func ToString(value interface{}, defVal ...string) string {
	if len(defVal) > 1 {
		panic("too many arguments")
	}

	var str string
	switch value.(type) {
	case string:
		str = value.(string)
	case []byte:
		str = string(value.([]byte))
	case int, int8, int16, int32, int64:
		str = strconv.FormatInt(reflect.ValueOf(value).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		str = strconv.FormatUint(reflect.ValueOf(value).Uint(), 10)
	case float32:
		str = strconv.FormatFloat(float64(value.(float32)), 'f', -1, 32)
	case float64:
		str = strconv.FormatFloat(value.(float64), 'f', -1, 64)
	default:
		str = ""
	}

	if str == "" && len(defVal) == 1 {
		return defVal[0]
	}
	return str
}

// 转换为int64类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToInt64(value interface{}, defVal ...int64) int64 {
	if len(defVal) > 1 {
		panic("too many arguments")
	}

	val := reflect.ValueOf(value)
	var d int64
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = val.Int()
	case uint, uint8, uint16, uint32, uint64:
		d = int64(val.Uint())
	case float32, float64:
		d = int64(val.Float())
	case string:
		d, err = strconv.ParseInt(val.String(), 10, 64)
	case []byte:
		d, err = strconv.ParseInt(string(value.([]byte)), 10, 64)
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}
	return d
}

// 转换为int类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToInt(value interface{}, defVal ...int) int {
	if len(defVal) > 1 {
		panic("too many arguments")
	}

	val := reflect.ValueOf(value)
	var d int
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = int(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = int(val.Uint())
	case float32, float64:
		d = int(val.Float())
	case string:
		d, err = strconv.Atoi(val.String())
	case []byte:
		d, err = strconv.Atoi(string(value.([]byte)))
	default:
		err = fmt.Errorf("ToInt64 need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}

	return d
}

// Iif 模拟三元运算符
func Iif(expr bool, trueVal interface{}, falseVal interface{}) interface{} {
	if expr {
		return trueVal
	}
	return falseVal
}
