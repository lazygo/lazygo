package mysql

import (
	"reflect"
	"strconv"
)

// toString 转换为string
func toString(value interface{}) string {
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
	case Raw:
		str = string(value.(Raw))
	default:
		str = ""
	}
	return str
}

// CreateAnyTypeSlice interface{}转为 []interface{}
func CreateAnyTypeSlice(slice interface{}) ([]interface{}, bool) {
	val, ok := isSlice(slice)
	if !ok {
		return nil, false
	}

	sliceLen := val.Len()
	out := make([]interface{}, sliceLen)
	for i := 0; i < sliceLen; i++ {
		out[i] = val.Index(i).Interface()
	}

	return out, true
}

// isSlice 判断是否为slice数据
func isSlice(arg interface{}) (reflect.Value, bool) {
	val := reflect.ValueOf(arg)
	ok := false
	if val.Kind() == reflect.Slice {
		ok = true
	}
	return val, ok
}

func mergeMap(maps ...map[string]interface{}) map[string]interface{} {
	var merged = make(map[string]interface{}, cap(maps))
	for _, m := range maps {
		for mk, mv := range m {
			merged[mk] = mv
		}
	}
	return merged
}
