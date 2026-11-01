package sqldb

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

// IN/NOT IN 参数去重：语义等价的前提下使查询更紧凑
func TestINDedup(t *testing.T) {
	assert := testify.New(t)

	// 两参形式（slice 值）去重并保持首次出现顺序
	builder := newBuilder(nil, "t")
	builder.Where("a", []int{1, 2, 1, 3, 2})
	sql, args, err := builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT `a` FROM `t` WHERE `a` IN(?, ?, ?)", sql)
	assert.Equal([]any{1, 2, 3}, args)

	// 显式 in 操作符
	builder = newBuilder(nil, "t")
	builder.Where("a", "in", []uint64{7, 7, 8})
	_, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal([]any{uint64(7), uint64(8)}, args)

	// NOT IN 同样去重
	builder = newBuilder(nil, "t")
	builder.WhereNotIn("a", []any{"x", "x", "y"})
	sql, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT `a` FROM `t` WHERE `a` NOT IN(?, ?)", sql)
	assert.Equal([]any{"x", "y"}, args)

	// map 条件中的 slice 值
	builder = newBuilder(nil, "t")
	builder.Where(map[string]any{"a": []int{5, 5}})
	_, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal([]any{5}, args)

	// 单元素与不重复列表原样
	builder = newBuilder(nil, "t")
	builder.Where("a", []int{9})
	_, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal([]any{9}, args)

	// 空列表语义不变（恒假）
	builder = newBuilder(nil, "t")
	builder.Where("a", []int{})
	sql, _, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Contains(sql, "1=0")

	// 不可比较元素不 panic、原样保留
	builder = newBuilder(nil, "t")
	builder.Where("a", "in", []any{[]string{"p"}, []string{"p"}})
	_, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Len(args, 2)

	// CONTAINS 不受影响（其参数是单值模糊匹配，且 "CONTAINS" 含子串 "IN"）
	builder = newBuilder(nil, "t")
	builder.Where("name", "contains", "in")
	sql, args, err = builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Contains(sql, "LIKE ?")
	assert.Len(args, 1)
}

// dedupINArgs 直接单测：边界与类型行为
func TestDedupINArgs(t *testing.T) {
	assert := testify.New(t)

	// 空与单元素原样
	assert.Nil(dedupINArgs(nil))
	one := []any{1}
	assert.Equal(one, dedupINArgs(one))

	// 保持首次出现顺序
	assert.Equal([]any{3, 1, 2}, dedupINArgs([]any{3, 1, 3, 2, 1}))

	// 全重复收敛为单个
	assert.Equal([]any{"x"}, dedupINArgs([]any{"x", "x", "x"}))

	// nil 元素参与去重（map[any] 支持 nil 键）
	assert.Equal([]any{nil, 1}, dedupINArgs([]any{nil, nil, 1}))

	// 跨类型同值不合并（SQL 参数类型不同，语义正确）
	mixed := dedupINArgs([]any{int(1), int64(1), int(1)})
	assert.Equal([]any{int(1), int64(1)}, mixed)

	// []byte 值相等但指针不同，保守保留不合并
	b1, b2 := []byte("ab"), []byte("ab")
	bs := dedupINArgs([]any{b1, b2})
	assert.Len(bs, 2)

	// 不可比较元素不 panic、原样保留
	nc := dedupINArgs([]any{[]string{"a"}, []string{"a"}, 7})
	assert.Len(nc, 3)
	assert.Equal(7, nc[2])
}

// IN 闸门：操作符拼写变体与占位符/参数对齐
func TestINOpGate(t *testing.T) {
	assert := testify.New(t)

	// 操作符大小写与空白变体均触发去重
	for _, op := range []string{"IN", "in", " In ", "NOT IN", "not in", "Not  IN"} {
		builder := newBuilder(nil, "t")
		builder.Where("a", op, []int{1, 1, 2})
		sql, args, err := builder.Select("a").QueryString()
		assert.Nil(err)
		assert.Contains(sql, "(?, ?)")
		assert.Equal([]any{1, 2}, args)
	}

	// map 条件中带操作符的 key（"a in" 走 key 解析分支）
	builder := newBuilder(nil, "t")
	builder.Where(map[string]any{"a in": []int{5, 5, 6}})
	_, args, err := builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Equal([]any{5, 6}, args)

	// 含 "IN" 子串的非列表操作符（BINARY 仅单参）不受影响
	builder = newBuilder(nil, "t")
	builder.Where("a", "BINARY", "v")
	sql, args, err := builder.Select("a").QueryString()
	assert.Nil(err)
	assert.Contains(sql, "BINARY(?)")
	assert.Equal([]any{"v"}, args)

	// CONTAINS 大小写变体不受影响
	for _, op := range []string{"contains", "CONTAINS"} {
		builder := newBuilder(nil, "t")
		builder.Where("name", op, "kw")
		_, args, err := builder.Select("a").QueryString()
		assert.Nil(err)
		assert.Len(args, 1)
	}
}
