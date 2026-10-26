package sqldb

// LikePattern 与 contains 操作符的测试（纯 SQL 形状断言，无需数据库）

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

func TestLikePattern(t *testing.T) {
	assert := testify.New(t)

	assert.Equal("%abc%", LikePattern("abc"))
	assert.Equal("%%", LikePattern("")) // 空关键字匹配全部
	// 通配符与转义符本身被转义
	assert.Equal(`%a\%b\_c\\d%`, LikePattern(`a%b_c\d`))
}

func TestContainsCond(t *testing.T) {
	assert := testify.New(t)

	// "field op", value 形式
	b := newBuilder(nil, "regress_check")
	sql, args, err := b.Where("name contains", "abc").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `name` LIKE ?", sql)
	assert.Equal([]any{"%abc%"}, args)

	// 关键字中的通配符被转义，不改变匹配行为
	b = newBuilder(nil, "regress_check")
	sql, args, err = b.Where("name contains", `a%b_c\d`).QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `name` LIKE ?", sql)
	assert.Equal([]any{`%a\%b\_c\\d%`}, args)

	// 三参数形式 Where(k, op, v)
	b = newBuilder(nil, "regress_check")
	sql, args, err = b.Where("title", "contains", "x").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `title` LIKE ?", sql)
	assert.Equal([]any{"%x%"}, args)

	// WhereMap 形式
	b = newBuilder(nil, "regress_check")
	sql, args, err = b.WhereMap(map[string]any{"name contains": "abc"}).QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `name` LIKE ?", sql)
	assert.Equal([]any{"%abc%"}, args)

	// contains 与其他条件组合
	b = newBuilder(nil, "regress_check")
	sql, args, err = b.Where("uid", 1).Where("name contains", "ab").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `uid` = ? AND `name` LIKE ?", sql)
	assert.Equal([]any{1, "%ab%"}, args)

	// like 操作符保持原样（向后兼容，由调用方自行处理模式）
	b = newBuilder(nil, "regress_check")
	sql, args, err = b.Where("name like", "%a%").QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE `name` like ?", sql)
	assert.Equal([]any{"%a%"}, args)
}

// 真实查询验证：contains 对通配符按字面量匹配
func TestContainsRealQuery(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "abc"})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "a%b"})
	db.Table("regress_check").Insert(map[string]any{"uid": 3, "name": "a_b"})
	db.Table("regress_check").Insert(map[string]any{"uid": 4, "name": "ab"})

	type row struct {
		UID  uint64 `json:"uid"`
		Name string `json:"name"`
	}

	// contains 中的 % 按字面量匹配，仅命中 a%b
	rows, _, err := db.Table("regress_check").Where("name contains", "a%b").Find[row]()
	assert.Nil(err)
	assert.Len(rows, 1)
	assert.Equal("a%b", rows[0].Name)

	// contains 中的 _ 按字面量匹配，仅命中 a_b
	rows, _, err = db.Table("regress_check").Where("name contains", "a_b").Find[row]()
	assert.Nil(err)
	assert.Len(rows, 1)
	assert.Equal("a_b", rows[0].Name)

	// 普通模糊匹配不受影响：命中 abc 和 ab
	rows, _, err = db.Table("regress_check").Where("name contains", "ab").Find[row]()
	assert.Nil(err)
	assert.Len(rows, 2)

	// like + LikePattern 与 contains 等价
	rows, _, err = db.Table("regress_check").Where("name like", LikePattern("a%b")).Find[row]()
	assert.Nil(err)
	assert.Len(rows, 1)
	assert.Equal("a%b", rows[0].Name)
}
