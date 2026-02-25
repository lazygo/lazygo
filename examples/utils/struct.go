// 结构体相关方法
package utils

import (
	"reflect"
	"slices"
)

// Struct2Map 将结构体转成map，除去在exclude中声明的tag
func Struct2Map[T any](obj T, exclude ...string) (map[string]any, map[string]any) {
	objValue := reflect.Indirect(reflect.ValueOf(obj))
	objType := objValue.Type()

	exclude = append(exclude, "-")
	result := make(map[string]any)
	excluded := make(map[string]any)
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		tag := field.Tag.Get("json")
		if tag != "" {
			if slices.Contains(exclude, tag) {
				excluded[tag] = objValue.Field(i).Interface()
			} else {
				result[tag] = objValue.Field(i).Interface()
			}
		}
	}
	return result, excluded
}
