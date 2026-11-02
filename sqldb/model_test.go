package sqldb

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

// defaultFields：JOIN 模型未显式指定字段时默认取全字段（含前缀别名），
// 修复 SELECT * 列名映射不到结构体导致静默返回全零数据的问题
// （线上事故：query_forward 回源返回 n>0 但字段全空的 EntryForwardData）
func TestDefaultFields(t *testing.T) {
	assert := testify.New(t)

	// JOIN 模型：默认字段含 E./F.（此处 O./U.）前缀别名
	fields := defaultFields[OrderWithUser](nil)
	assert.NotEmpty(fields)
	assert.Contains(fields, "`O`.`name` AS `O.name`")
	assert.Contains(fields, "`U`.`uid` AS `U.uid`")
	assert.Contains(fields, "`U`.`vip`.`deadline` AS `U.vip.deadline`")

	// 显式指定的字段原样透传
	explicit := []string{"`O`.`id`"}
	assert.Equal(explicit, defaultFields[OrderWithUser](explicit))

	// 非结构体类型回退默认行为（fields 保持为空 → 上层按 * 处理）
	assert.Nil(defaultFields[int](nil))
}

// JOIN 模型的 SELECT 语句必须生成带前缀别名的列，而不是 SELECT *
//（SELECT * 的列名无法映射回 E/F 前缀结构体，扫描结果全零且无任何报错）
func TestJoinModelSelectSQL(t *testing.T) {
	assert := testify.New(t)

	join := Table(Table("core_entry").As("E").LeftJoin(Table("core_forward").As("F"), "E.forward_id", "F.id").String())
	cond := map[string]any{"E.node_id": 5, "E.entry": "demo"}

	// 旧行为：不指定字段 → SELECT *，列名映射失败（复现线上 query_forward 空数据）
	before, _, err := newBuilder(nil, join).Where(cond).Select().QueryString()
	assert.Nil(err)
	assert.Contains(before, "SELECT *")

	// 新行为：TxModel.First/Find 经 defaultFields 生成显式别名列
	after, _, err := newBuilder(nil, join).Where(cond).Select(defaultFields[OrderWithUser](nil)...).QueryString()
	assert.Nil(err)
	assert.NotContains(after, "SELECT *")
	assert.Contains(after, "`O`.`name` AS `O.name`")
	assert.Contains(after, "`U`.`uid` AS `U.uid`")
}
