package sqldb

// 已修复问题的回归测试

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	testify "github.com/stretchr/testify/assert"
)

type regressData struct {
	Id    uint64 `json:"id"`
	Uid   uint64 `json:"uid"`
	Name  string `json:"name"`
	Age   int64  `json:"age"`
	Ctime int64  `json:"ctime"`
	Mtime int64  `json:"mtime"`
}

type regressModel struct {
	TxModel[regressData]
}

func regressDB(t *testing.T) *DB {
	assert := testify.New(t)
	err := Init([]MySQLConfig{
		{
			Driver:          "mysql",
			Name:            "unit_test_reg",
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
	db, err := Database("unit_test_reg")
	assert.Nil(err)
	db.Exec("DROP TABLE IF EXISTS regress_check")
	_, err = db.Exec("CREATE TABLE `regress_check` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`uid` bigint(20) unsigned NOT NULL DEFAULT 0," +
		"`name` varchar(64) NOT NULL DEFAULT ''," +
		"`age` int(10) unsigned NOT NULL DEFAULT 0," +
		"`nick` varchar(64) DEFAULT NULL," +
		"`t` datetime DEFAULT NULL," +
		"`ctime` int(10) unsigned NOT NULL DEFAULT 0," +
		"`mtime` int(10) unsigned NOT NULL DEFAULT 0," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	assert.Nil(err)
	return db
}

// One() 带别名/表达式
func TestRegressOne(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a"})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "b"})

	c, err := db.Table("regress_check").Where("id", ">", 0).One("count(*) AS c")
	assert.Nil(err)
	assert.Equal("2", c)

	name, err := db.Table("regress_check").Where("uid", 1).One("name")
	assert.Nil(err)
	assert.Equal("a", name)

	// 反引号包裹的字段名
	name, err = db.Table("regress_check").Where("uid", 1).One("`name`")
	assert.Nil(err)
	assert.Equal("a", name)
}

// 空切片 IN / NOT IN
func TestRegressEmptyIn(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a"})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "b"})

	sql, _, err := db.Table("regress_check").WhereIn("id", []any{}).QueryString()
	assert.Nil(err)
	assert.Contains(sql, "1=0")

	sql, _, err = db.Table("regress_check").WhereNotIn("id", []any{}).QueryString()
	assert.Nil(err)
	assert.Contains(sql, "1=1")

	// NOT IN 空集应命中全部行
	_, n, err := db.Table("regress_check").WhereNotIn("id", []any{}).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(2, n)

	// IN 空集应命中 0 行
	_, n, err = db.Table("regress_check").WhereIn("id", []any{}).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(0, n)
}

// nil 值条件
func TestRegressNilCond(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "nick": nil})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "nick": "x"})

	var rows []map[string]any
	rows, n, err := db.Table("regress_check").Where("nick", nil).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal("1", rows[0]["uid"])

	rows, n, err = db.Table("regress_check").Where("nick", "!=", nil).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal("2", rows[0]["uid"])

	// WhereMap 中的 nil
	rows, n, err = db.Table("regress_check").WhereMap(map[string]any{"nick": nil}).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(1, n)
}

// Modify 自定义主键
func TestRegressModify(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	mdl := &regressModel{}
	mdl.SetTable("regress_check")
	mdl.SetTx(&db.Tx)

	id, err := mdl.Create(&regressData{Uid: 100, Name: "origin"})
	assert.Nil(err)

	// 按 uid 更新
	n, err := mdl.Modify(&regressData{Id: uint64(id), Uid: 100, Name: "renamed"}, "uid")
	assert.Nil(err)
	assert.Equal(int64(1), n)

	got, cnt, err := mdl.First(map[string]any{"uid": 100})
	assert.Nil(err)
	assert.Equal(1, cnt)
	assert.Equal("renamed", got.Name)

	// 默认按 id 更新
	n, err = mdl.Modify(&regressData{Id: uint64(id), Uid: 100, Name: "byid"})
	assert.Nil(err)
	assert.Equal(int64(1), n)

	// 主键不存在时返回错误
	_, err = mdl.Modify(&regressData{Id: uint64(id)}, "nonexistent")
	assert.Equal(ErrInvalidArguments, err)
}

// Insert 不再改写调用方的 map
func TestRegressInsertClone(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	mdl := &regressModel{}
	mdl.SetTable("regress_check")
	mdl.SetTx(&db.Tx)

	set := map[string]any{"uid": 1, "name": "a"}
	_, err := mdl.Insert(set)
	assert.Nil(err)
	_, hasCtime := set["ctime"]
	_, hasMtime := set["mtime"]
	assert.False(hasCtime)
	assert.False(hasMtime)
}

// Paginator 分组统计
func TestRegressPaginatorGroupBy(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a", "age": 1})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "b", "age": 1})
	db.Table("regress_check").Insert(map[string]any{"uid": 3, "name": "c", "age": 2})
	db.Table("regress_check").Insert(map[string]any{"uid": 4, "name": "d", "age": 2})
	db.Table("regress_check").Insert(map[string]any{"uid": 5, "name": "e", "age": 2})

	mdl := &regressModel{}
	mdl.SetTable("regress_check")
	mdl.SetTx(&db.Tx)

	// 不传 fields 的分组分页：默认查询分组字段，且 total 为组数
	resp, n, err := mdl.Paginator(map[string]any{}, &PaginatorRequest{Page: 1, Size: 10, Order: "age", Sort: "asc"}, []string{"age"})
	assert.Nil(err)
	assert.Equal(2, n)
	assert.Equal(int64(2), resp.Total)

	// 非分组分页 total 为行数
	resp, n, err = mdl.Paginator(map[string]any{}, &PaginatorRequest{Page: 1, Size: 10, Order: "id", Sort: "asc"}, nil)
	assert.Nil(err)
	assert.Equal(5, n)
	assert.Equal(int64(5), resp.Total)
}

// First 空结果返回零值, 0, nil
func TestRegressFirstEmpty(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	mdl := &regressModel{}
	mdl.SetTable("regress_check")
	mdl.SetTx(&db.Tx)

	data, n, err := mdl.First(map[string]any{"uid": 404})
	assert.Nil(err)
	assert.Equal(0, n)
	assert.Zero(data)
}

// 写操作后 Builder 状态清空
func TestRegressBuilderClear(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a", "age": 1})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "b", "age": 2})

	b := db.Table("regress_check")
	n, err := b.Where("uid", 1).OrderBy("age", "desc").Limit(5).Update(map[string]any{"name": "a1"})
	assert.Nil(err)
	assert.Equal(int64(1), n)

	sql, _, err := b.Where("uid", 2).QueryString()
	assert.Nil(err)
	assert.NotContains(sql, "ORDER BY")
	assert.NotContains(sql, "LIMIT")

	// Delete 后同样清空
	n, err = b.Where("uid", 2).GroupBy("age").Delete()
	assert.Nil(err)
	assert.Equal(int64(1), n)
	sql, _, err = b.Where("uid", 1).QueryString()
	assert.Nil(err)
	assert.NotContains(sql, "GROUP BY")
}

// BeforeHook：读路径生效，写路径不泄漏
func TestRegressBeforeHook(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a"})

	calls := 0
	hook := func(q string, a ...any) func() {
		calls++
		return func() {}
	}

	_, _, err := db.Table("regress_check").BeforeHook(hook).Where("id", ">", 0).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(1, calls)

	// 写路径
	n, err := db.Table("regress_check").BeforeHook(hook).Where("uid", 1).Update(map[string]any{"name": "b"})
	assert.Nil(err)
	assert.Equal(int64(1), n)
	assert.Equal(2, calls)

	// 后续普通查询不再触发 hook
	_, _, err = db.Table("regress_check").Where("id", ">", 0).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(2, calls)

	// Count 路径也生效
	_, err = db.Table("regress_check").BeforeHook(hook).Where("id", ">", 0).Count()
	assert.Nil(err)
	assert.Equal(3, calls)
}

// 并发下 BeforeHook 无数据竞争（需配合 -race 运行）
func TestRegressBeforeHookRace(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "a"})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hook := func(q string, a ...any) func() { return func() {} }
			for j := 0; j < 30; j++ {
				_, err := db.Table("regress_check").BeforeHook(hook).Where("uid", 1).Update(map[string]any{"name": "b"})
				assert.Nil(err)
				_, _, err = db.Table("regress_check").BeforeHook(hook).Where("id", ">", 0).Find[map[string]any]()
				assert.Nil(err)
			}
		}()
	}
	wg.Wait()
}

// json tag 选项与 time.Time 字段扫描
func TestRegressParser(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	now := time.Now().Truncate(time.Second)
	db.Table("regress_check").Insert(map[string]any{"uid": 7, "t": now})

	type withTag struct {
		Id  uint64    `json:"id"`
		Uid uint64    `json:"uid,omitempty"`
		T   time.Time `json:"t"`
	}

	var row withTag
	row, _, err := db.Table("regress_check").Where("uid", 7).First[withTag]()
	assert.Nil(err)
	assert.Equal(uint64(7), row.Uid)
	assert.Equal(now.UTC(), row.T.UTC())

	// time.Time 按叶子字段处理
	arr := getFieldArr(reflect.ValueOf(&withTag{}).Elem())
	assert.Contains(mapKeysOf(arr), "t")
	assert.Contains(mapKeysOf(arr), "uid")

	// StructFields 对非结构体入参不再 panic
	assert.Nil(StructFields(1))
}

// Transaction panic 保留 error 类型，且正常回滚
func TestRegressTransaction(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	// panic(error)
	sentinel := errors.New("boom")
	err := db.Transaction(func(tx *Tx) error {
		tx.Table("regress_check").Insert(map[string]any{"uid": 1})
		panic(sentinel)
	})
	assert.ErrorIs(err, sentinel)

	n, err := db.Table("regress_check").Where("uid", 1).Count()
	assert.Nil(err)
	assert.Equal(int64(0), n)

	// panic(非error)
	err = db.Transaction(func(tx *Tx) error {
		panic("plain string")
	})
	assert.NotNil(err)
	assert.Contains(err.Error(), "plain string")

	// 返回错误时回滚
	err = db.Transaction(func(tx *Tx) error {
		tx.Table("regress_check").Insert(map[string]any{"uid": 2})
		return sentinel
	})
	assert.ErrorIs(err, sentinel)
	n, err = db.Table("regress_check").Where("uid", 2).Count()
	assert.Nil(err)
	assert.Equal(int64(0), n)
}

// WhereRaw 拼接 orderBy/limit 组合仍正常（防御 Clear 影响）
func TestRegressQueryString(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	b := db.Table("regress_check")
	sql, args, err := b.WhereRaw("uid IN(?, ?)", 1, 2).Where("name", "a").OrderBy("id", "desc").Limit(3).Offset(6).QueryString()
	assert.Nil(err)
	assert.Equal("SELECT * FROM `regress_check` WHERE uid IN(?, ?) AND `name` = ? ORDER BY `id` DESC LIMIT 6, 3", sql)
	assert.Equal([]any{1, 2, "a"}, args)
}

// sqlx 风格：key 中携带运算符（"field op", value）
func TestRegressFieldOpCond(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check")

	db.Table("regress_check").Insert(map[string]any{"uid": 1, "name": "abc", "age": 20})
	db.Table("regress_check").Insert(map[string]any{"uid": 2, "name": "xyz", "age": 30})

	var check = func(cond ...any) (string, []any) {
		b := newBuilder(nil, "regress_check")
		sql, args, err := b.Where(cond...).QueryString()
		assert.Nil(err)
		return sql, args
	}

	sql, args := check("age >", 18)
	assert.Equal("SELECT * FROM `regress_check` WHERE `age` > ?", sql)
	assert.Equal([]any{18}, args)

	sql, _ = check("age >=", 18)
	assert.Equal("SELECT * FROM `regress_check` WHERE `age` >= ?", sql)

	sql, _ = check("name like", "%a%")
	assert.Equal("SELECT * FROM `regress_check` WHERE `name` like ?", sql)

	sql, args = check("uid in", []any{1, 2})
	assert.Equal("SELECT * FROM `regress_check` WHERE `uid` IN(?, ?)", sql)
	assert.Equal([]any{1, 2}, args)

	sql, _ = check("deleted_at IS", nil)
	assert.Equal("SELECT * FROM `regress_check` WHERE `deleted_at` IS NULL", sql)

	// 联表字段名带点号
	sql, args = check("U.age >", 18)
	assert.Equal("SELECT * FROM `regress_check` WHERE `U`.`age` > ?", sql)
	assert.Equal([]any{18}, args)

	// 实际查询验证
	var rows []map[string]any
	rows, n, err := db.Table("regress_check").Where("age >", 25).Find[map[string]any]()
	assert.Nil(err)
	assert.Equal(1, n)
	assert.Equal("2", rows[0]["uid"])
}

// AutoCTime/AutoMTime：指定字段名后自动维护创建/更新时间，默认不写入
func TestRegressModelTimeField(t *testing.T) {
	assert := testify.New(t)
	db := regressDB(t)
	defer db.Exec("DROP TABLE IF EXISTS regress_check_time")
	_, err := db.Exec("CREATE TABLE `regress_check_time` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT," +
		"`name` varchar(64) NOT NULL DEFAULT ''," +
		"`created_at` int(10) unsigned NOT NULL DEFAULT 0," +
		"`updated_at` int(10) unsigned NOT NULL DEFAULT 0," +
		"PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	assert.Nil(err)

	type timeData struct {
		Id        uint64 `json:"id"`
		Name      string `json:"name"`
		CreatedAt int64  `json:"created_at"`
		UpdatedAt int64  `json:"updated_at"`
	}
	type timeModel struct {
		TxModel[timeData]
	}
	newModel := func() *timeModel {
		m := &timeModel{}
		m.SetTable("regress_check_time")
		m.SetTx(&db.Tx)
		return m
	}

	// 默认不自动写入
	mdl := newModel()
	id, err := mdl.Create(&timeData{Name: "a"})
	assert.Nil(err)
	data, _, err := mdl.First(map[string]any{"id": id})
	assert.Nil(err)
	assert.Equal(int64(0), data.CreatedAt)
	assert.Equal(int64(0), data.UpdatedAt)

	// 指定字段名后：Create写入创建时间和更新时间
	mdl = newModel()
	mdl.AutoCTime("created_at")
	mdl.AutoMTime("updated_at")
	now := time.Now().Unix()
	id, err = mdl.Create(&timeData{Name: "b"})
	assert.Nil(err)
	data, _, err = mdl.First(map[string]any{"id": id})
	assert.Nil(err)
	assert.GreaterOrEqual(data.CreatedAt, now-2)
	assert.GreaterOrEqual(data.UpdatedAt, now-2)

	// Modify 更新更新时间字段
	n, err := mdl.Modify(&timeData{Id: uint64(id), Name: "c", CreatedAt: data.CreatedAt})
	assert.Nil(err)
	assert.Equal(int64(1), n)
	data, _, err = mdl.First(map[string]any{"id": id})
	assert.Nil(err)
	assert.GreaterOrEqual(data.UpdatedAt, now-2)
}

func mapKeysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
