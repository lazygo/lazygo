package mysql

import (
	"reflect"
	"slices"
	"testing"

	testify "github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
)

type User struct {
	UID     uint64 `json:"uid"`
	Name    string `json:"name"`
	VipInfo `json:"vip"`
}

type VipInfo struct {
	UID      uint64 `json:"uid"`
	Name     string `json:"name"`
	Ctime    uint64 `json:"ctime"`
	Deadline uint64 `json:"deadline"`
}

type Order struct {
	ID   uint64 `json:"id"`
	UID  string `json:"uid"`
	Name string `json:"name"`
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
	fielsArr := StructFields(rv)

	keys := maps.Keys(fielsArr)
	slices.Sort(keys)

	testData := []string{"hh", "O.id", "O.uid", "O.name", "U.vip.uid", "U.vip.name", "U.vip.ctime", "U.vip.deadline", "U.uid", "U.name"}
	slices.Sort(testData)
	assert.Equal(keys, testData)
}
