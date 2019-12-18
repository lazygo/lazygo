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

type Manager struct {
	databases map[string]*Db
	conf      *gjson.Result
	lock      *sync.RWMutex
}

/*
[
	{"name": "test", "user": "root", "passwd": "root", "host": "127.0.0.1", "port": 3306, "dbname": "test", "charset": "utf8", "maxOpenConns": 1000},
	{"name": "test1", "user": "root", "passwd": "root", "host": "127.0.0.1", "port": 3306, "dbname": "test1", "charset": "utf8", "maxOpenConns": 1000}
]
*/

func NewManager(conf *gjson.Result) (*Manager, error) {
	m := &Manager{
		databases: map[string]*Db{},
		conf:      conf,
		lock:      new(sync.RWMutex),
	}
	for _, item := range conf.Array() {
		err := m.connect(&item)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) connect(item *gjson.Result) error {

	user := item.Get("user").String()
	passwd := item.Get("passwd").String()
	host := item.Get("host").String()
	port := item.Get("port").Int()
	name := item.Get("name").String()
	dbname := item.Get("dbname").String()
	charset := item.Get("charset").String()
	maxOpenConns := item.Get("maxOpenConns").Int()

	dataSourceName := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=5s",
		DefaultString(user, "root"),
		DefaultString(passwd, "root"),
		DefaultString(host, "127.0.0.1"),
		DefaultInt64(port, 3306),
		DefaultString(dbname, "test"),
		DefaultString(charset, "utf8"),
	)
	database, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	database.SetMaxIdleConns(16)
	database.SetMaxOpenConns(int(DefaultInt64(maxOpenConns, 1000)))
	database.SetConnMaxLifetime(time.Duration(60) * time.Second)
	m.lock.Lock()
	m.databases[name] = NewDb(name, database)
	m.lock.Unlock()
	return nil
}

func (m *Manager) Database(name string) (*Db, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if databases, ok := m.databases[name]; ok {
		return databases, nil
	}
	return nil, errors.New("指定数据库不存在")
}
