package locker

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

func randomToken() uint64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Uint64()
}

func randRange(min uint64, max uint64) uint64 {
	if min > max {
		max, min = min, max
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return min + uint64(r.Int63n(int64(max - min)))
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

// 转换为int64
func toInt64(value interface{}) int64 {
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
	return d
}
