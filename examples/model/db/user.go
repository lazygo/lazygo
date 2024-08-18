package db

import (
	"database/sql"
	"time"

	"github.com/lazygo/lazygo/examples/model"
)

type UserModel struct {
	model.DBModel[UserData]
}

type UserData struct {
	UID      uint64         `json:"uid"`
	Email    sql.NullString `json:"email"`
	Mobile   sql.NullString `json:"mobile"`
	Password string         `json:"password"`
	Avator   string         `json:"avator"`
	CTime    uint64         `json:"ctime"`
	MTime    uint64         `json:"mtime"`
}

func NewUserModel() *UserModel {
	mdl := &UserModel{}
	mdl.SetTable("user")
	mdl.SetDb("lazygo-db")
	return mdl
}

func (mdl *UserModel) FetchByUid(uid uint64, fields ...string) (*UserData, int, error) {
	cond := map[string]any{
		"uid": uid,
	}
	return mdl.First(cond, fields...)
}

func (mdl *UserModel) UpdateByUid(uid int64, data map[string]any) (int64, error) {
	cond := map[string]any{
		"uid": uid,
	}
	data["mtime"] = time.Now().Unix()
	return mdl.QueryBuilder().Where(cond).Update(data)
}
