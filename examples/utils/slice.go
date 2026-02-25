// slice 的常见操作
package utils

import (
	"fmt"
	"strings"
)

// 将某个结构体的slice,转换成指定主键的map
// 如 [{id:1,name:2}, {id:2,name:3}] 转换成 {1:{id:1,name:2}, 2:{id:2,name:3}}
// 可用于将数据库的查询结果转成map,便于后续使用
// 注意这里的primarykey是结构体字段的名字而不是tag
func Slice2Map[T any, K comparable](objs []T, first bool, primaryKey func(T) K) map[K]T {
	var result = make(map[K]T, len(objs))
	for _, item := range objs {
		k := primaryKey(item)
		if _, ok := result[k]; first && ok {
			continue
		}
		result[k] = item
	}
	return result
}

func Slice2KV[T any, K comparable, V any](objs []T, first bool, primaryKey func(T, int) (K, V)) map[K]V {
	var result = make(map[K]V, len(objs))
	for index, item := range objs {
		k, v := primaryKey(item, index)
		if _, ok := result[k]; first && ok {
			continue
		}
		result[k] = v
	}
	return result
}

// SliceIndex 将切片转换成以主键为key的map,值为切片
func SliceIndex[T any, K comparable, V any](objs []T, primaryKey func(T) (K, V)) map[K][]V {
	var result = make(map[K][]V)
	for _, item := range objs {
		k, v := primaryKey(item)
		result[k] = append(result[k], v)
	}
	return result
}

// SliceColumns 获取切片每个元素中的指定值组成的新切片
func SliceColumns[T any, S comparable](objs []T, primaryKey func(T) S) []S {
	var m = make([]S, 0, len(objs))
	for _, item := range objs {
		m = append(m, primaryKey(item))
	}
	return m
}

// SliceFilter 过滤切片
func SliceFilter[T any](objs []T, filter func(T) bool) []T {
	var m = make([]T, 0, len(objs))
	for _, item := range objs {
		if filter(item) {
			m = append(m, item)
		}
	}
	return m
}

func SliceUnique[T any, S comparable](objs []T, primaryKey func(T) S) []S {
	var m = make([]S, 0, len(objs))
	seen := make(map[S]struct{}, len(objs))
	for _, item := range objs {
		val := primaryKey(item)
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		m = append(m, val)
	}
	return m
}

// SliceJoin 将任意数值类型的slice切片转换成 用 , 拼接起来的字符串
func SliceJoin[T any](list []T, separator string) string {
	var s []string
	for _, item := range list {
		s = append(s, fmt.Sprintf("%v", item))
	}
	return strings.Join(s, separator)
}
