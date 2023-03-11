package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// ToString 转换为字符串
// 如果value不能转换为字符串，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为空字符串""
func ToString(value interface{}, defVal ...string) string {
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

// ToInt64 转换为int64类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToInt64(value interface{}, defVal ...int64) int64 {
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

// ToUint64 转换为uint64类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToUint64(value interface{}, defVal ...uint64) uint64 {
	val := reflect.ValueOf(value)
	var d uint64
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = uint64(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = val.Uint()
	case float32, float64:
		d = uint64(val.Float())
	case string:
		d, err = strconv.ParseUint(val.String(), 10, 64)
	case []byte:
		d, err = strconv.ParseUint(string(value.([]byte)), 10, 64)
	default:
		err = fmt.Errorf("ToUint64 need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}
	return d
}

// ToInt 转换为int类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToInt(value interface{}, defVal ...int) int {
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
		err = fmt.Errorf("ToInt need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}

	return d
}

// ToUint 转换为uint类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToUint(value interface{}, defVal ...uint) uint {
	val := reflect.ValueOf(value)
	var d uint
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = uint(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = uint(val.Uint())
	case float32, float64:
		d = uint(val.Float())
	case string:
		var d64 uint64
		d64, err = strconv.ParseUint(val.String(), 10, 32)
		d = uint(d64)
	case []byte:
		var d64 uint64
		d64, err = strconv.ParseUint(string(value.([]byte)), 10, 32)
		d = uint(d64)
	default:
		err = fmt.Errorf("ToUint need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}

	return d
}

// ToFloat 转换为float64类型
// value 可以是数字、数字字符串
// 如果value不能转换为数字，则返回默认值defVal
// defVal 提供默认值，如果没有，则视为0
func ToFloat(value interface{}, defVal ...float64) float64 {
	val := reflect.ValueOf(value)
	var d float64
	var err error
	switch value.(type) {
	case int, int8, int16, int32, int64:
		d = float64(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = float64(val.Uint())
	case float32, float64:
		d = val.Float()
	case string:
		d, err = strconv.ParseFloat(val.String(), 64)
	case []byte:
		d, err = strconv.ParseFloat(string(value.([]byte)), 64)
	default:
		err = fmt.Errorf("ToUint need numeric not `%T`", value)
	}
	if err != nil {
		d = 0
	}
	if d == 0 && len(defVal) == 1 {
		return defVal[0]
	}

	return d
}
