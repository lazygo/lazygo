// 实现 bool
package utils

import (
	"reflect"
)

// Bool 真值测试
func Bool(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return v != 0
	case uint, uint8, uint16, uint32, uint64:
		return v != 0
	case float32, float64:
		return v != 0
	case string:
		return v != ""
	case []any:
		return len(v) != 0
	default:
		return reflect.ValueOf(value).Len() > 0
	}
}
