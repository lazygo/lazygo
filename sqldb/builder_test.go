package sqldb

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestWhere(t *testing.T) {
	assert := testify.New(t)

	builder := newBuilder(nil, "table_name")

	arr := []any{1, 2}

	builder.
		Where("a", 1).
		Where("b", "=", 2).
		Where("c=3").
		Where("d", arr).
		Where(map[string]any{"e": 1, "f": "ff", "g": arr}).
		WhereNotIn("h", arr).
		WhereRaw("(x=? OR y=?)", "aa", 99).
		Where("table_name.field", 1)

	data := []string{
		"last_view_time",
		"last_view_user",
		"table_name.field",
	}
	sql, args, err := builder.Select(data...).QueryString()
	assert.Nil(err, "err.Error()")

	sp := strings.Split(sql, " WHERE ")
	assert.Equal(len(sp), 2, "错误")
	assert.Equal(sp[0], "SELECT `last_view_time`, `last_view_user`, `table_name`.`field` FROM `table_name`", "错误")

	cond := sp[1]
	expectedCond := strings.Split(cond, " AND ")
	for k := range expectedCond {
		for strings.Count(expectedCond[k], "?") > 0 && len(args) > 0 {
			expectedCond[k] = strings.Replace(expectedCond[k], "?", fmt.Sprintf("%v", args[0]), 1)
			args = args[1:]
		}
	}
	sort.Strings(expectedCond)

	assert.Equal(len(expectedCond), 10, "错误")
	assert.Equal(expectedCond[0], "(x=aa OR y=99)", "错误")
	assert.Equal(expectedCond[1], "`a` = 1", "错误")
	assert.Equal(expectedCond[2], "`b` = 2", "错误")
	assert.Equal(expectedCond[3], "`d` IN(1, 2)", "错误")
	assert.Equal(expectedCond[4], "`e` = 1", "错误")
	assert.Equal(expectedCond[5], "`f` = ff", "错误")
	assert.Equal(expectedCond[6], "`g` IN(1, 2)", "错误")
	assert.Equal(expectedCond[7], "`h` NOT IN(1, 2)", "错误")
	assert.Equal(expectedCond[8], "`table_name`.`field` = 1", "错误")
	assert.Equal(expectedCond[9], "c=3", "错误")

	// ((a=1 AND b=2) OR c=3) AND d=4
	// map 遍历顺序随机，以下 8 种排列（2x2x2）均为合法结果，
	// 校验时必须保证 SQL 与 args 的位置一一对应
	// QueryString 不改变构建器状态，这里使用新的 Builder，避免与上面的用例叠加条件
	builder2 := newBuilder(nil, "table_name")
	builder2.Where(map[string]any{
		string(OR): map[string]any{
			string(AND): map[string]any{
				"a": 1,
				"b": 2,
			},
			"c": 3,
		},
		"d": 4,
	})
	sql, args, err = builder2.Select(data...).QueryString()
	assert.Nil(err, "err.Error()")

	sp = strings.Split(sql, " WHERE ")
	assert.Equal(len(sp), 2, "错误")

	type expect struct {
		cond string
		args []any
	}
	expected := []expect{
		{"((`a` = ? AND `b` = ?) OR `c` = ?) AND `d` = ?", []any{1, 2, 3, 4}},
		{"((`b` = ? AND `a` = ?) OR `c` = ?) AND `d` = ?", []any{2, 1, 3, 4}},
		{"(`c` = ? OR (`a` = ? AND `b` = ?)) AND `d` = ?", []any{3, 1, 2, 4}},
		{"(`c` = ? OR (`b` = ? AND `a` = ?)) AND `d` = ?", []any{3, 2, 1, 4}},
		{"`d` = ? AND ((`a` = ? AND `b` = ?) OR `c` = ?)", []any{4, 1, 2, 3}},
		{"`d` = ? AND ((`b` = ? AND `a` = ?) OR `c` = ?)", []any{4, 2, 1, 3}},
		{"`d` = ? AND (`c` = ? OR (`a` = ? AND `b` = ?))", []any{4, 3, 1, 2}},
		{"`d` = ? AND (`c` = ? OR (`b` = ? AND `a` = ?))", []any{4, 3, 2, 1}},
	}
	matched := false
	for _, e := range expected {
		if sp[1] == e.cond && testify.ObjectsAreEqual(e.args, args) {
			matched = true
			break
		}
	}
	assert.True(matched, "WHERE 与 args 不匹配: %s args=%v", sp[1], args)

}

func TestBuildVal(t *testing.T) {
	set := map[string]any{
		"last_view_time": "12345678",
		"last_view_user": "li",
	}
	extra := []string{
		"view_num=view_num+1",
	}
	strset, args := buildVal(set, extra)

	arrSet := strings.Split(strset, ", ")
	for k := range arrSet {
		for strings.Count(arrSet[k], "?") > 0 && len(args) > 0 {
			arrSet[k] = strings.Replace(arrSet[k], "?", fmt.Sprintf("%v", args[0]), 1)
			args = args[1:]
		}
	}
	sort.Strings(arrSet)
	strset = strings.Join(arrSet, ", ")

	assert := testify.New(t)
	assert.Equal(strset, "`last_view_time` = 12345678, `last_view_user` = li, view_num=view_num+1", "错误")

}
