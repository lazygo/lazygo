package sqldb

// JOIN 模型查询默认字段的回归测试。
// 复现线上事故：First/Find 未显式指定字段时生成 SELECT *，
// 列名映射不到 E/F 前缀结构体，返回 n>0 但字段全零的数据。

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	testify "github.com/stretchr/testify/assert"
)

type joinEntryItem struct {
	Id        uint64 `json:"id"`
	NodeID    uint64 `json:"node_id"`
	Scheme    string `json:"scheme"`
	Entry     string `json:"entry"`
	ForwardID uint64 `json:"forward_id"`
}

type joinForwardItem struct {
	Id     uint64 `json:"id"`
	PeerID uint64 `json:"peer_id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
}

type joinEntryForwardData struct {
	joinEntryItem   `json:"E"`
	joinForwardItem `json:"F"`
}

type joinEntryForwardModel struct {
	TxModel[joinEntryForwardData]
}

// 本地库自建表与数据，测试后清理
func joinModelDB(t *testing.T) (*DB, *joinEntryForwardModel) {
	assert := testify.New(t)
	err := Init([]MySQLConfig{
		{
			Driver:          "mysql",
			Name:            "unit_test_join",
			User:            "root",
			Passwd:          "",
			Host:            "127.0.0.1",
			Port:            3306,
			DbName:          "unit_test",
			MaxOpenConns:    10,
			MaxIdleConns:    10,
			ConnMaxLifetime: 60,
			Params:          map[string]string{"charset": "utf8mb4", "timeout": "5s", "parseTime": "true"},
		},
	})
	assert.Nil(err)
	db, err := Database("unit_test_join")
	assert.Nil(err)

	db.Exec("DROP TABLE IF EXISTS join_entry")
	db.Exec("DROP TABLE IF EXISTS join_forward")
	_, err = db.Exec("CREATE TABLE `join_forward` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`peer_id` bigint(20) unsigned NOT NULL DEFAULT 0," +
		"`status` int(10) NOT NULL DEFAULT 0," +
		"`title` varchar(64) NOT NULL DEFAULT ''," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	assert.Nil(err)
	_, err = db.Exec("CREATE TABLE `join_entry` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`node_id` bigint(20) unsigned NOT NULL DEFAULT 0," +
		"`scheme` varchar(16) NOT NULL DEFAULT ''," +
		"`entry` varchar(64) NOT NULL DEFAULT ''," +
		"`forward_id` bigint(20) unsigned NOT NULL DEFAULT 0," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	assert.Nil(err)

	mdl := &joinEntryForwardModel{}
	table := Table("join_entry").As("E").LeftJoin(Table("join_forward").As("F"), "E.forward_id", "F.id")
	mdl.SetTable(table.String())
	err = mdl.SetDB("unit_test_join")
	assert.Nil(err)

	// 一条暂停态(status=2)的转发规则及其入口，与线上 query_forward 场景一致
	forwardID, err := db.Table("join_forward").Insert(map[string]any{"peer_id": 83, "status": 2, "title": "demo-forward"})
	assert.Nil(err)
	_, err = db.Table("join_entry").Insert(map[string]any{"node_id": 5, "scheme": "http", "entry": "demo-entry", "forward_id": forwardID})
	assert.Nil(err)

	t.Cleanup(func() {
		db.Exec("DROP TABLE IF EXISTS join_entry")
		db.Exec("DROP TABLE IF EXISTS join_forward")
	})
	return db, mdl
}

func TestJoinModelFirstDefaultFields(t *testing.T) {
	assert := testify.New(t)
	_, mdl := joinModelDB(t)

	cond := map[string]any{"E.node_id": 5, "E.scheme": "http", "E.entry": "demo-entry"}

	// 未指定字段：默认全字段（含 E./F. 别名），两侧结构体都正确映射。
	// 修复前此处 n=1 但 F.Status=0（线上 query_forward 空数据事故）
	data, n, err := mdl.First(cond)
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal(2, data.Status)
	assert.Equal(uint64(83), data.PeerID)
	assert.Equal("demo-forward", data.Title)
	assert.Equal("demo-entry", data.Entry)
	assert.True(data.ForwardID > 0)

	// 显式 "*" 与未指定同义，同样展开为模型字段
	star, n, err := mdl.First(cond, "*")
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal(2, star.Status)
	assert.Equal("demo-entry", star.Entry)

	// Find 同路径
	list, n, err := mdl.Find(cond)
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal(2, list[0].Status)
	assert.Equal("demo-forward", list[0].Title)

	// Paginator 同路径
	page, n, err := mdl.Paginator(cond, &PaginatorRequest{Page: 1, Size: 10}, nil)
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal(1, len(page.List))
	assert.Equal(2, page.List[0].Status)
}

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

	// 显式 "*"（含反引号包裹）与未指定同义，展开为模型字段
	assert.Contains(defaultFields[OrderWithUser]([]string{"*"}), "`O`.`name` AS `O.name`")
	assert.Contains(defaultFields[OrderWithUser]([]string{"`*`"}), "`U`.`uid` AS `U.uid`")

	// 非结构体类型回退默认行为（fields 保持为空 → 上层按 * 处理）
	assert.Nil(defaultFields[int](nil))
}
