package db

import (
	"database/sql"
	"time"

	"github.com/lazygo-dev/lazygo/examples/framework"
	"github.com/lazygo-dev/lazygo/examples/model"
	"github.com/lazygo/lazygo/mysql"
)

const (
	AppidMain = 1 // 主应用
)

// UserModel 用户
type UserModel struct {
	Ctx framework.Context
	model.TxModel[UserData]
}

type UserData struct {
	UID      uint64         `json:"uid"`
	Appid    int            `json:"appid"`
	Email    sql.NullString `json:"email"`
	Mobile   sql.NullString `json:"mobile"`
	Password string         `json:"password"`
	CTime    int64          `json:"ctime"`
	MTime    int64          `json:"mtime"`
}

func NewUserModel(ctx framework.Context) *UserModel {
	mdl := &UserModel{Ctx: ctx}
	mdl.SetTable("uc_user")
	mdl.SetDB("lazygo-db")
	return mdl
}

// Create 创建用户
func (mdl *UserModel) Create(user map[string]any, trans ...func(*mysql.Tx, uint64) error) (uid uint64, err error) {
	user["ctime"] = time.Now().Unix()
	err = model.DB("zj").Transaction(func(tx *mysql.Tx) error {
		id, err := tx.Table("uc_user").Insert(user)
		if err != nil {
			return err
		}
		uid = uint64(id)

		for _, fn := range trans {
			if fn == nil {
				continue
			}
			if err := fn(tx, uid); err != nil {
				return err
			}
		}
		return nil
	})
	return uid, err
}

func (mdl *UserModel) FetchByUid(uid uint64, fields ...string) (*UserData, int, error) {
	cond := map[string]any{
		"uid": uid,
	}
	data, n, err := mdl.First(cond, fields...)
	if err != nil {
		return nil, 0, err
	}

	return data, n, err
}

func (mdl *UserModel) UpdateByUid(uid uint64, data map[string]any) (int64, error) {
	cond := map[string]any{
		"uid": uid,
	}
	data["mtime"] = time.Now().Unix()
	return mdl.QueryBuilder().Where(cond).Update(data)
}
