package mysql

import (
	"database/sql"
	"reflect"
	"slices"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

type User struct {
	UID  uint64 `json:"uid"`
	Name string `json:"name"`
	Vip  `json:"vip"`
}

type Vip struct {
	UID      uint64 `json:"uid"`
	Name     string `json:"name"`
	Ctime    uint64 `json:"ctime"`
	Deadline uint64 `json:"deadline"`
}

type Order struct {
	ID   uint64           `json:"id"`
	UID  string           `json:"uid"`
	Hh   string           `json:"-"`
	Name sql.Null[string] `json:"name"`
}

type OrderWithUser struct {
	HH    string `json:"hh"`
	User  `json:"U"`
	Order `json:"O"`
}

func TestGetfieldArr(t *testing.T) {
	assert := testify.New(t)

	var data OrderWithUser
	rv := reflect.ValueOf(&data).Elem()
	fielsArr := getFieldArr(rv)

	keys := maps.Keys(fielsArr)
	slices.Sort(keys)

	testData := []string{"hh", "O.id", "O.uid", "O.name", "U.vip.uid", "U.vip.name", "U.vip.ctime", "U.vip.deadline", "U.uid", "U.name"}
	slices.Sort(testData)
	assert.Equal(keys, testData)
}

func TestStructFields(t *testing.T) {
	assert := testify.New(t)

	var data OrderWithUser
	keys := StructFields(data, "O.id")
	slices.Sort(keys)

	testData := []string{"`O`.`name` AS `O.name`", "`O`.`uid` AS `O.uid`", "`U`.`name` AS `U.name`", "`U`.`uid` AS `U.uid`", "`U`.`vip`.`ctime` AS `U.vip.ctime`", "`U`.`vip`.`deadline` AS `U.vip.deadline`", "`U`.`vip`.`name` AS `U.vip.name`", "`U`.`vip`.`uid` AS `U.vip.uid`", "`hh`"}
	slices.Sort(testData)
	assert.Equal(keys, testData)
}
