package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tidwall/gjson"
	"sync"
	"time"
)

type mysqlManager struct {
	databases map[string]*Db
	conf      *gjson.Result
	lock      *sync.RWMutex
}

// 单例
var manager *mysqlManager = nil

/*
配置文件格式
[
	{"name": "test_w", "user": "root", "passwd": "root", "host": "127.0.0.1", "port": 3306, "dbname": "test", "charset": "utf8", "maxOpenConns": 1000},
	{"name": "test1-r", "user": "root", "passwd": "root", "host": "127.0.0.1", "port": 3306, "dbname": "test1", "charset": "utf8", "maxOpenConns": 1000}
]
*/
// 框架初始化时调用，请不要在业务中调用
// Init 初始化数据库连接
func Init(conf *gjson.Result) error {
	if manager != nil {
		panic("数据库不能重复初始化")
	}
	// 保持单例
	manager = &mysqlManager{
		databases: map[string]*Db{},
		conf:      conf,
		lock:      new(sync.RWMutex),
	}
	for _, item := range conf.Array() {
		err := manager.connect(&item)
		if err != nil {
			return err
		}
	}
	return nil
}

// 连接数据库
func (m *mysqlManager) connect(item *gjson.Result) error {
	name := item.Get("name").String()
	m.lock.RLock()
	_, ok := m.databases[name]
	m.lock.RUnlock()
	if ok {
		// 已连接的就不再次连接了
		return nil
	}
	user := item.Get("user").String()
	passwd := item.Get("passwd").String()
	host := item.Get("host").String()
	port := item.Get("port").Int()
	dbname := item.Get("dbname").String()
	charset := item.Get("charset").String()
	prefix := item.Get("prefix").String()
	maxOpenConns := item.Get("maxOpenConns").Int()
	maxIdleConns := item.Get("maxIdleConns").Int()
	connMaxLifetime := item.Get("connMaxLifetime").Int()
	dataSourceName := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=5s",
		defaultString(user, "root"),
		defaultString(passwd, "root"),
		defaultString(host, "127.0.0.1"),
		defaultInt64(port, 3306),
		defaultString(dbname, "test"),
		defaultString(charset, "utf8"),
	)
	database, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	database.SetMaxIdleConns(int(defaultInt64(maxIdleConns, 16)))
	database.SetMaxOpenConns(int(defaultInt64(maxOpenConns, 1000)))
	if connMaxLifetime > 0 {
		database.SetConnMaxLifetime(time.Duration(defaultInt64(connMaxLifetime, 60)) * time.Second)
	}
	m.lock.Lock()
	m.databases[name] = newDb(name, database, prefix)
	m.lock.Unlock()
	return nil
}

// Database 通过名称获取数据库
func Database(name string) (*Db, error) {
	if manager == nil {
		return nil, errors.New("未初始化Mysql")
	}
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	if databases, ok := manager.databases[name]; ok {
		return databases, nil
	}
	return nil, errors.New("指定数据库不存在，或未初始化")
}
