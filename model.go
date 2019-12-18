package lazygo

import (
	"github.com/lazygo/lazygo/memcache"
	"github.com/lazygo/lazygo/mysql"
)

type Model struct {
}

func (m *Model) Db(name string) *mysql.Db {
	return App().GetDb(name)
}

func (m *Model) Mc(name string) *memcache.Memcache {
	return App().GetMc(name)
}
