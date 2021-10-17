package mysql

import (
	"strings"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
)

type UserInfo struct {
	Id    int    `json:"id"`
	Uid   uint64 `json:"uid"`
	Name  string `json:"name"`
	Ctime int64  `json:"ctime"`
}

func TestDb(t *testing.T) {
	assert := testify.New(t)

	conf := []*Config{
		{
			Name:            "unit_test",
			User:            "root",
			Passwd:          "root",
			Host:            "127.0.0.1",
			Port:            3306,
			DbName:          "abc",
			Charset:         "utf8mb4",
			Prefix:          "good_",
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

	tableName := "unit_test"

	tableSql := "CREATE TABLE `good_unit_test` (" +
		"`id` int(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key'," +
		"`uid` bigint(20) unsigned NOT NULL COMMENT 'uid'," +
		"`name` varchar(20) NOT NULL DEFAULT '' COMMENT 'name'," +
		"`ctime` int(10) unsigned NOT NULL DEFAULT 0 COMMENT 'create time'," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY `uniq_uid` (`uid`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='good_unit_test';"

	dropSql := "DROP TABLE `good_unit_test`;"

	now := time.Now().Unix()

	db, err := Database("unit_test")
	assert.Nil(err)

	_, err = db.Exec(tableSql)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	defer func() {
		_, err = db.Exec(dropSql)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		ok, err := CheckTable(db, "good_"+tableName)
		if err != nil {
			assert.Nil(err, err.Error())
		}
		assert.False(ok)
	}()

	ok, err := CheckTable(db, "good_"+tableName)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.True(ok)

	// Test Insert
	data := map[string]interface{}{
		"uid":   1001,
		"name":  "测试",
		"ctime": now,
	}
	id, err := db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	// Test FetchRowIn
	user := &UserInfo{}
	err = db.Table(tableName).Where("id", id).FetchRowIn([]interface{}{"uid", "name", "ctime"}, user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Uid, uint64(1001))
	assert.Equal(user.Name, "测试")
	assert.Equal(user.Ctime, now)

	// Test Update
	set := map[string]interface{}{
		"name": "goodcoder",
	}
	n, err := db.Table(tableName).Where("uid", "=", user.Uid).Update(set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	name, err := db.Table(tableName).Where("uid", user.Uid).FetchOne("name")
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
	name, err = db.Table(tableName).Where("uid", user.Uid).FetchOne("name")
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(name, "UpdateRaw Test")

	// Test Update Without Cond
	_, err = db.Table(tableName).Update(set)
	assert.NotNil(err)

	// Test Increment
	set = map[string]interface{}{
		"name": "小明",
	}
	n, err = db.Table(tableName).Where("id", id).Increment("ctime", 1, set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	user = &UserInfo{}
	err = db.Table(tableName).Where("id", id).FetchRowIn([]interface{}{"name", "ctime"}, user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Name, "小明")
	assert.Equal(user.Ctime, now+1)

	// Test Decrement
	set = map[string]interface{}{
		"name": "goodcoder",
	}
	n, err = db.Table(tableName).Where("id", id).Decrement("ctime", 1, set)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))
	user = &UserInfo{}
	err = db.Table(tableName).Where("id", id).FetchRowIn([]interface{}{"name", "ctime"}, user)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(user.Name, "goodcoder")
	assert.Equal(user.Ctime, now)

	// Test FetchRow
	result, err := db.Table(tableName).Where("id", id).FetchRow([]interface{}{"uid", "name"})
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(result["uid"], "1001")
	assert.Equal(result["name"], "goodcoder")

	// Test FetchOne
	name, err = db.Table(tableName).Where("id", id).FetchOne("name")
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(name, "goodcoder")

	// Test FetchAll
	data = map[string]interface{}{
		"uid":   1002,
		"name":  "测试2号",
		"ctime": now,
	}
	_, err = db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	list, err := db.Table(tableName).Where("id", ">", 0).Fetch([]interface{}{"uid", "name"})
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(len(list), 2)
	assert.Equal(list[0]["uid"], "1001")
	assert.Equal(list[0]["name"], "goodcoder")
	assert.Equal(list[1]["uid"], "1002")
	assert.Equal(list[1]["name"], "测试2号")

	// Test FetchAllIn
	var listSS []UserInfo
	err = db.Table(tableName).Where("id", ">", 0).FetchIn([]interface{}{"uid", "name"}, &listSS)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(len(listSS), 2)
	assert.Equal(listSS[0].Uid, uint64(1001))
	assert.Equal(listSS[0].Name, "goodcoder")
	assert.Equal(listSS[1].Uid, uint64(1002))
	assert.Equal(listSS[1].Name, "测试2号")

	// Test FetchWithPage
	data = map[string]interface{}{
		"uid":   1003,
		"name":  "测试3号",
		"ctime": now,
	}
	_, err = db.Table(tableName).Insert(data)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	dataList, err := db.Table(tableName).Where("id", ">", 0).FetchWithPage([]interface{}{"uid", "name"}, 1, 2)
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
	cond := map[string]interface{}{
		"uid": in,
	}
	dataList, err = db.Table(tableName).Where(cond).FetchWithPage([]interface{}{"uid", "name"}, 2, 2)
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(dataList.Count, int64(3))
	assert.Equal(len(dataList.List), 1)
	assert.Equal(dataList.List[0]["uid"], "1003")

	dataMap := dataList.ToMap()
	assert.Equal(dataMap["count"], int64(3))
	lst, ok := dataMap["list"].([]map[string]interface{})
	assert.True(ok)
	assert.Equal(len(lst), 1)
	assert.Equal(lst[0]["uid"], "1003")

	// Test Delete
	cond = map[string]interface{}{
		"id": id,
	}
	n, err = db.Table(tableName).Where(cond).Delete()
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Equal(n, int64(1))

	result, err = db.Table(tableName).Where("id", id).FetchRow([]interface{}{"uid", "name"})
	if err != nil {
		assert.Nil(err, err.Error())
	}
	assert.Empty(result)

	// test GetTablePrefix
	prefix := db.GetTablePrefix()
	assert.Equal(prefix, conf[0].Prefix)
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
	result, err := parseData(rows)
	if err != nil {
		return false, err
	}
	if len(result) == 0 {
		return false, nil
	}
	return true, nil
}
