package lazygo

import (
	"github.com/lazygo/lazygo/memcache"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/utils"
)

type Model struct {
}

func (m *Model) Db(name string) *mysql.Db {
	db, err :=  mysql.Database(name)
	utils.CheckError(err)
	return db
}

func (m *Model) Mc(name string) *memcache.Memcache {
	mc, err := memcache.Mc(name)
	utils.CheckError(err)
	return mc
}
