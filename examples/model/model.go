package model

import (
	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/mysql"
)

type DbModel struct {
	table  string
	dbname string
}

func (m *DbModel) SetTable(table string) {
	m.table = table
}

func (m *DbModel) SetDb(dbname string) {
	m.dbname = dbname
}

func (m *DbModel) Db(dbName string) *mysql.DB {
	database, err := mysql.Database(dbName)
	if err != nil {
		panic(err)
	}
	return database
}

func (m *DbModel) QueryBuilder() mysql.Builder {
	table := m.table
	if table == "" {
		// 没有指定表名
		panic("没有指定表名")
	}
	database, err := mysql.Database(m.dbname)
	if err != nil {
		panic("数据库不存在")
	}
	return database.Table(table)
}

type CacheModel struct {
	Cache cache.Cache
}

func (m *CacheModel) SetCache(name string) {
	m.Cache = m.cache(name)
}

func (m *CacheModel) cache(name string) cache.Cache {
	instance, err := cache.Instance(name)
	if err != nil {
		panic(err)
	}
	return instance
}
