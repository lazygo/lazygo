package library

import (
	"fmt"
	"reflect"
	"strconv"
)

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
	case int:
		str = strconv.Itoa(value.(int))
	case int64:
		str = strconv.FormatInt(value.(int64), 10)
	default:
		str = ""
	}

	if str == "" && len(defVal) == 1 {
		return defVal[0]
	}
	return str
}

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
	case string:
		d, err = strconv.ParseInt(val.String(), 10, 64)
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
	case string:
		d, err = strconv.Atoi(val.String())
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
