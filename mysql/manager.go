package mysql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Name            string `json:"name" toml:"name"`
	User            string `json:"user" toml:"user"`
	Passwd          string `json:"passwd" toml:"passwd"`
	Host            string `json:"host" toml:"host"`
	Port            int    `json:"port" toml:"port"`
	DbName          string `json:"dbname" toml:"dbname"`
	Charset         string `json:"charset" toml:"charset"`
	Prefix          string `json:"prefix" toml:"prefix"`
	MaxOpenConns    int    `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

type Manager struct {
	sync.Map
}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config) error {
	for _, item := range conf {
		if _, ok := manager.Load(item.Name); ok {
			// 已连接的就不再次连接了
			continue
		}
		db, err := manager.open(item)
		if err != nil {
			return err
		}
		err = db.Ping()
		if err != nil {
			return err
		}
		m.Store(item.Name, newDb(item.Name, db, item.Prefix))
	}
	return nil
}

// closeAll 关闭数据库连接
func (m *Manager) closeAll() error {
	var err error
	m.Range(func(name, db interface{}) bool {
		err = db.(*DB).Close()
		if err != nil {
			return false
		}
		m.Delete(name)
		return true
	})
	return err
}

// open 连接数据库
func (m *Manager) open(item Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=5s",
		item.User,
		item.Passwd,
		item.Host,
		item.Port,
		item.DbName,
		item.Charset,
	)
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(item.MaxIdleConns)
	database.SetMaxOpenConns(item.MaxOpenConns)
	if item.ConnMaxLifetime > 0 {
		database.SetConnMaxLifetime(time.Duration(item.ConnMaxLifetime) * time.Second)
	}
	return database, nil
}

// Init 初始化数据库
func Init(conf []Config) error {
	return manager.init(conf)
}

// CloseAll 关闭数据库连接
func CloseAll() error {
	return manager.closeAll()
}

// Database 通过名称获取数据库
func Database(name string) (*DB, error) {
	if databases, ok := manager.Load(name); ok {
		return databases.(*DB), nil
	}
	return nil, ErrDatabaseNotExists
}
