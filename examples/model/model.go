package model

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/httpclient"
	"github.com/lazygo/lazygo/mysql"
)

type DBModel[T any] struct {
	table  string
	dbname string
}

func (m *DBModel[T]) SetTable(table string) {
	m.table = table
}

func (m *DBModel[T]) SetDb(dbname string) {
	m.dbname = dbname
}

func (m *DBModel[T]) DB(dbName string) *mysql.DB {
	database, err := mysql.Database(dbName)
	if err != nil {
		panic(err)
	}
	return database
}

func (m *DBModel[T]) QueryBuilder() mysql.Builder {
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

func (mdl *DBModel[T]) First(cond map[string]any, fields ...string) (*T, int, error) {
	var data T
	n, err := mdl.QueryBuilder().Where(cond).Select(fields...).First(&data)
	if err != nil {
		return nil, 0, err
	}
	return &data, n, nil
}

func (mdl *DBModel[T]) Exists(cond map[string]any) (bool, error) {
	var data map[string]any
	n, err := mdl.QueryBuilder().Where(cond).Select("(0)").First(&data)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (mdl *DBModel[T]) Find(cond map[string]any, fields ...string) ([]T, int, error) {
	var data []T
	n, err := mdl.QueryBuilder().Where(cond).Select(fields...).Find(&data)
	if err != nil {
		return nil, 0, err
	}
	return data, n, nil
}

func (mdl *DBModel[T]) Insert(set map[string]any) (int64, error) {
	set["ctime"] = time.Now().Unix()
	return mdl.QueryBuilder().Insert(set)
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
		panic(fmt.Sprintln(name, err))
	}
	return instance
}

type RpcModel struct {
	Client  *httpclient.Client
	BaseURL string
	IPP     []netip.AddrPort
}

func DefaultClient() *httpclient.Client {
	return httpclient.New(&httpclient.Config{
		// DNSResolverAddr:"",
		HTTPDNSAdapter: "baidu",
	}).Client(30 * time.Second)
}
