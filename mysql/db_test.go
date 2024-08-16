package mysql

import (
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
)

type UserInfo struct {
	Id     int    `json:"id"`
	Uid    uint64 `json:"uid"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
	Ctime  int64  `json:"ctime"`
}

type VipInfo struct {
	Id       int    `json:"id"`
	Uid      uint64 `json:"uid"`
	Name     string `json:"name"`
	Level    int64  `json:"level"`
	Deadline int64  `json:"deadline"`
	Ctime    int64  `json:"ctime"`
}

type UserVipInfo struct {
	UserInfo
	VipInfo
}

type UserVipInfoWithAlias struct {
	UserInfo `json:"U"`
	VipInfo  `json:"V"`
}

func TestDb(t *testing.T) {
	assert := testify.New(t)

	conf := []Config{
		{
			Name:            "unit_test",
			User:            "root",
			Passwd:          "",
			Host:            "127.0.0.1",
			Port:            3306,
			DbName:          "unit_test",
			Charset:         "utf8mb4",
			MaxOpenConns:    10,
			MaxIdleConns:    10,
			ConnMaxLifetime: 60,
		},
	}
	err := Init(conf)
	assert.Nil(err)

	defer func() {
		err := CloseAll()
		assert.Nil(err)
		_, err = Database("unit_test")
		assert.NotNil(err)
	}()

	tableName := "good_unit_test"
	tableVipName := "good_unit_test_vip"

	tableSql := "CREATE TABLE `good_unit_test` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key'," +
		"`uid` bigint(20) unsigned NOT NULL COMMENT 'uid'," +
		"`name` varchar(20) NOT NULL DEFAULT '' COMMENT 'name'," +
		"`mobile` varchar(20) NOT NULL DEFAULT '' COMMENT 'mobile'," +
		"`ctime` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'create time'," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY `uniq_uid` (`uid`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='good_unit_test';"

	tableVipSql := "CREATE TABLE `good_unit_test_vip` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key'," +
		"`uid` bigint(20) unsigned NOT NULL COMMENT 'uid'," +
		"`name` varchar(20) NOT NULL DEFAULT '' COMMENT 'name'," +
		"`level` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'level'," +
		"`deadline` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'deadline'," +
		"`ctime` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'create time'," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY `uniq_uid` (`uid`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='good_unit_test_vip';"

	dropSql := "DROP TABLE `good_unit_test`;"
	dropVipSql := "DROP TABLE `good_unit_test_vip`;"

	now := time.Now().Unix()

	db, err := Database("unit_test")
	assert.Nil(err)

	_, err = db.Exec(tableSql)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	_, err = db.Exec(tableVipSql)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	defer func() {
		_, err = db.Exec(dropSql)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		ok, err := CheckTable(db, tableName)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		assert.False(ok)
		_, err = db.Exec(dropVipSql)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		ok, err = CheckTable(db, tableVipName)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		assert.False(ok)
	}()

	ok, err := CheckTable(db, tableName)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.True(ok)

	// Test Insert
	data := map[string]any{
		"uid":    1001,
		"name":   "测试",
		"mobile": "13812345678",
		"ctime":  now,
	}
	uid, err := db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	data = map[string]any{
		"uid":      1001,
		"name":     "测试vip",
		"level":    1,
		"deadline": now + 1000000,
		"ctime":    now,
	}
	id, err := db.Table(tableVipName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	// Test FetchRowIn
	user := &UserInfo{}
	_, err = db.Table(tableName).Where("id", uid).Select("uid", "name", "ctime").First(user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Uid, uint64(1001))
	assert.Equal(user.Name, "测试")
	assert.Equal(user.Ctime, now)

	// Test Join Table
	uservip := &UserVipInfo{}
	table := Table(tableName).As("U").LeftJoin(Table(tableVipName).As("V"), "U.uid", "V.uid")
	_, err = db.Table(table.String()).
		Where("U.uid", 1001).First(uservip)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Uid, uint64(1001))
	assert.Equal(user.Name, "测试")
	assert.Equal(user.Ctime, now)

	uservip2 := &UserVipInfoWithAlias{}
	_, err = db.Table(table.String()).
		Where("U.uid", 1001).Select(StructFields(uservip2)...).First(uservip2)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Uid, uint64(1001))
	assert.Equal(user.Name, "测试")
	assert.Equal(user.Ctime, now)
	sql, args, err := db.Table(table.String()).
		Where("U.uid", 1001).Select(StructFields(uservip2)...).QueryString()
	assert.Nil(err, nil)
	assert.Equal(sql, "SELECT `U`.`ctime` AS `U.ctime`, `U`.`id` AS `U.id`, `U`.`mobile` AS `U.mobile`, `U`.`name` AS `U.name`, `U`.`uid` AS `U.uid`, `V`.`ctime` AS `V.ctime`, `V`.`deadline` AS `V.deadline`, `V`.`id` AS `V.id`, `V`.`level` AS `V.level`, `V`.`name` AS `V.name`, `V`.`uid` AS `V.uid` FROM `good_unit_test` AS `U` LEFT JOIN `good_unit_test_vip` AS `V` ON `U`.`uid` = `V`.`uid` WHERE `U`.`uid` = ?")
	assert.Equal(args, []any{1001})

	// Test Update
	set := map[string]any{
		"name": "goodcoder",
	}
	n, err := db.Table(tableName).Where("uid", "=", user.Uid).Update(set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	name, err := db.Table(tableName).Where("uid", user.Uid).One("name")
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(name, "goodcoder")

	// Test UpdateRaw
	strSet := "name='UpdateRaw Test'"
	n, err = db.Table(tableName).Where("uid", "=", user.Uid).UpdateRaw(strSet, 1)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	name, err = db.Table(tableName).Where("uid", user.Uid).One("name")
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(name, "UpdateRaw Test")

	// Test Update Without Cond
	_, err = db.Table(tableName).Update(set)
	assert.NotNil(err)

	// Test Increment
	set = map[string]any{
		"name": "小明",
	}
	n, err = db.Table(tableName).Where("id", id).Increment("ctime", 1, set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	user = &UserInfo{}
	_, err = db.Table(tableName).Where("id", id).Select("name", "ctime").First(user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Name, "小明")
	assert.Equal(user.Ctime, now+1)

	// Test Decrement
	set = map[string]any{
		"name": "goodcoder",
	}
	n, err = db.Table(tableName).Where("id", id).Decrement("ctime", 1, set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	user = &UserInfo{}
	_, err = db.Table(tableName).Where("id", id).Select("name", "ctime").First(user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Name, "goodcoder")
	assert.Equal(user.Ctime, now)

	// Test FetchOne
	name, err = db.Table(tableName).Where("id", id).One("name")
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(name, "goodcoder")

	// Test FetchAll
	data = map[string]any{
		"uid":   1002,
		"name":  "测试2号",
		"ctime": now,
	}
	_, err = db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	var listSS []UserInfo
	_, err = db.Table(tableName).Where("id", ">", 0).Select("uid", "name").Find(&listSS)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(len(listSS), 2)
	if len(listSS) >= 1 {
		assert.Equal(listSS[0].Uid, uint64(1001))
		assert.Equal(listSS[0].Name, "goodcoder")
	}
	if len(listSS) >= 2 {
		assert.Equal(listSS[1].Uid, uint64(1002))
		assert.Equal(listSS[1].Name, "测试2号")
	}
	_, err = db.Table(tableName).Where("id", ">", 0).Select("uid", "name").Find(&listSS)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(len(listSS), 2)

	// Fetch Map
	var userMap []map[string]any
	_, err = db.Table(tableName).Where("id", ">", 0).Select("uid", "name", "ctime").Find(&userMap)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(len(userMap), 2)
	if len(userMap) >= 1 {
		assert.Equal(userMap[0]["uid"], "1001")
		assert.Equal(userMap[0]["name"], "goodcoder")
	}
	if len(userMap) >= 2 {
		assert.Equal(userMap[1]["uid"], "1002")
		assert.Equal(userMap[1]["name"], "测试2号")
	}

	// Test FetchWithPage
	data = map[string]any{
		"uid":   1003,
		"name":  "测试3号",
		"ctime": now,
	}
	_, err = db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	dataList, err := db.Table(tableName).Where("id", ">", 0).Select("uid", "name").FetchWithPage(1, 2)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(dataList.Count, int64(3))
	assert.Equal(len(dataList.List), 2)
	assert.Equal(dataList.List[0]["uid"], "1001")
	assert.Equal(dataList.List[1]["uid"], "1002")

	in := []int{
		1001, 1002, 1003,
	}
	cond := map[string]any{
		"uid": in,
	}
	dataList, err = db.Table(tableName).Where(cond).Select("uid", "name").FetchWithPage(2, 2)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(dataList.Count, int64(3))
	assert.Equal(len(dataList.List), 1)
	if len(dataList.List) >= 1 {
		assert.Equal(dataList.List[0]["uid"], "1003")
	}

	dataMap := dataList.ToMap()
	assert.Equal(dataMap["count"], int64(3))
	lst, ok := dataMap["list"].([]map[string]any)
	assert.True(ok)
	assert.Equal(len(lst), 1)
	if len(lst) >= 1 {
		assert.Equal(lst[0]["uid"], "1003")
	}

	// Test Delete
	cond = map[string]any{
		"id": id,
	}
	n, err = db.Table(tableName).Where(cond).Delete()
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))

	var result []map[string]any
	_, err = db.Table(tableName).Where("id", id).Select("uid", "name").Find(&result)
	if err != nil {
		assert.Nil(err, ErrInvalidResultPtr)
	}
	assert.Empty(result)

	// test error
	_, err = db.Table(tableName).Where("id", ">", id).Select("nofound1", "name").First(&result)
	assert.True(strings.Contains(err.(*SqlError).Unwrap().Error(), "Unknown column 'nofound1' in 'field list'"))

	// test error
	n, err = db.Table(tableName).Update(map[string]any{"name1": 1})
	assert.Equal(err, ErrEmptyCond)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Delete(1)
	assert.Equal(err, ErrEmptyCond)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Where("1=1").Update(map[string]any{})
	assert.Equal(err, ErrEmptyValue)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Where("1=1").UpdateRaw("")
	assert.Equal(err, ErrEmptyValue)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Where(1, 2, 3).UpdateRaw("132")
	assert.Equal(err, ErrInvalidCondArguments)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Insert(map[string]any{})
	assert.Equal(err, ErrEmptyValue)
	assert.Equal(n, int64(0))

	n, err = db.Table(tableName).Where("1, 2, 3").UpdateRaw("132", 12, 21)
	assert.Equal(err, ErrInvalidArguments)
	assert.Equal(n, int64(0))

	x, err := db.Table(tableName).Where("1=1").OrderBy("1", "2").One("a")
	assert.Equal(err, ErrInvalidArguments)
	assert.Equal(x, "")

	x, err = db.Table(tableName).Where("1=1").GroupBy("id").One("id")
	assert.Equal(err, nil)
	assert.Equal(x, "2")

	n, err = db.Table(tableName).Where("1, 2, 3").Increment("132", 12, map[string]any{}, map[string]any{})
	assert.Equal(err, ErrInvalidArguments)
	assert.Equal(n, int64(0))

	resErrType := 0
	_, err = db.Table(tableName).Where("id", ">", id).Select("id", "name").First(&resErrType)
	assert.Equal(err, ErrInvalidResultPtr)

	var resNilType map[string]any
	_, err = db.Table(tableName).Where("id", ">", id).Select("id", "name").First(&resNilType)
	assert.Equal(err, ErrInvalidResultPtr)

	var resErrType2 []map[string]int
	_, err = db.Table(tableName).Where("id", ">", id).Select("id", "name").Find(&resErrType2)
	assert.Equal(err, ErrInvalidResultPtr)

}

func CheckTable(db *DB, table string) (bool, error) {
	descTable := "DESC " + table + ";"
	rows, err := db.Query(descTable)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return false, nil
		}
		return false, err
	}
	defer rows.Close()
	var result []map[string]any
	_, err = parseData(rows, &result)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	return true, nil
}
