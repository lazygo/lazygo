package db

import (
	"time"

	"github.com/lazygo/lazygo/examples/model"
)

type UserModel struct {
	model.DbModel
}

type UserData struct {
	UID      uint64 `json:"uid"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
	Avator   string `json:"avator"`
	CTime    uint64 `json:"ctime"`
	MTime    uint64 `json:"mtime"`
}

func NewUserModel() *UserModel {
	mdl := &UserModel{}
	mdl.SetTable("user")
	mdl.SetDb("lazygo-db")
	return mdl
}

func (mdl *UserModel) Create(data map[string]interface{}) (int64, error) {
	data["ctime"] = time.Now().Unix()
	return mdl.QueryBuilder().Insert(data)
}

func (mdl *UserModel) FetchByUid(fields []interface{}, uid int64) (*UserData, int, error) {
	cond := map[string]interface{}{
		"uid": uid,
	}
	return mdl.FetchRow(fields, cond)
}

func (mdl *UserModel) FetchRow(fields []interface{}, cond map[string]interface{}) (*UserData, int, error) {
	var data UserData
	n, err := mdl.QueryBuilder().Where(cond).FetchRow(fields, &data)
	if err != nil {
		return nil, 0, err
	}
	return &data, n, nil
}

func (mdl *UserModel) Exists(cond map[string]interface{}) (bool, error) {
	fields := []interface{}{"uid"}
	var data UserData
	n, err := mdl.QueryBuilder().Where(cond).FetchRow(fields, &data)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (mdl *UserModel) UpdateByUid(uid int64, data map[string]interface{}) (int64, error) {
	cond := map[string]interface{}{
		"uid": uid,
	}
	data["mtime"] = time.Now().Unix()
	return mdl.QueryBuilder().Where(cond).Update(data)
}
