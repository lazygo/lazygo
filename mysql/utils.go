package mysql

import (
	"reflect"
	"strings"
)

// CreateAnyTypeSlice interface{}转为 []interface{}
func CreateAnyTypeSlice(slice any) ([]any, bool) {
	val, ok := isSlice(slice)
	if !ok {
		return nil, false
	}

	sliceLen := val.Len()
	out := make([]any, sliceLen)
	for i := 0; i < sliceLen; i++ {
		out[i] = val.Index(i).Interface()
	}

	return out, true
}

// isSlice 判断是否为slice数据
func isSlice(arg any) (reflect.Value, bool) {
	val := reflect.ValueOf(arg)
	ok := false
	if val.Kind() == reflect.Slice {
		ok = true
	}
	return val, ok
}

func mergeMap(maps ...map[string]any) map[string]any {
	var merged = make(map[string]any, cap(maps))
	for _, m := range maps {
		for mk, mv := range m {
			merged[mk] = mv
		}
	}
	return merged
}

// buildVal 构建值
// val map[string]interface{} 将会被展开成 k=? 的形式 多个元素用逗号隔开
// extra []string 多个元素间用逗号隔开，并追加在 val参数展开的字符串后面
// 示例：val = {"last_view_time": "12345678", "last_view_user": "li"}; extra = ["view_num=view_num+1"]
func buildVal(val map[string]any, extra []string) (string, []any) {
	var items []string
	var args []any

	// 使用等号连接map的key和value，并放入数组
	for k, v := range val {
		items = append(items, build(k, "="))
		args = append(args, v)
	}

	// 追加额外的数组元素
	for _, v := range extra {
		v = strings.Trim(v, " ")
		if v != "" {
			items = append(items, v)
		}
	}
	// 用逗号将数组元素拼接成字符串
	result := strings.Join(items, ", ")
	return result, args
}

// build 构造
// 用操作符op把key和value占位符连接起来
// 示例 k = "name", op = "="    `name`=?
func build(k string, op string) string {
	return buildK(k) + " " + op + " ?"
}

// buildK key增加反引号
func buildK(k string) string {
	k = strings.TrimSpace(k)
	if !isSimple(k) {
		return k
	}
	return "`" + strings.ReplaceAll(k, ".", "`.`") + "`"
}

type Fields []string

func (f Fields) String() string {
	l := len(f)
	if l == 0 {
		return "*"
	}
	arr := make([]string, l, cap(f))
	for i, v := range f {
		v = strings.TrimSpace(v)
		arr[i] = buildK(v)
	}
	return strings.Join(arr, ", ")
}

func isSimple(v string) bool {
	return !strings.ContainsAny(strings.TrimSpace(v), "() `*/+-%=&<>!")
}

// ResultData 分页返回数据 - 返回结果定义
type ResultData struct {
	List      []map[string]any `json:"list"`
	Count     int64            `json:"count"`
	PageSize  int64            `json:"page_size"`
	Page      int64            `json:"page"`
	PageCount int64            `json:"page_count"`
	Start     int64            `json:"start"`
	Mark      int64            `json:"mark"`
}

// ToMap 分页结果集转化为map
func (r *ResultData) ToMap() map[string]any {
	if r == nil {
		return map[string]any{}
	}
	return map[string]any{
		"list":       r.List,
		"count":      r.Count,
		"page_size":  r.PageSize,
		"page":       r.Page,
		"page_count": r.PageCount,
		"start":      r.Start,
		"mark":       r.Mark,
	}
}
