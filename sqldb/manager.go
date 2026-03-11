package sqldb

import (
	"database/sql"
	"io"
	"sync"
	"time"
)

type Manager struct {
	sync.Map
}

var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []Config) error {
	for _, item := range conf {
		if _, ok := manager.Load(item.name()); ok {
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
		m.Store(item.name(), &DB{Tx{
			invoker: db,
			name:    item.name(),
			before:  func(query string, args ...any) func() { return func() {} },
		}})
	}
	return nil
}

// closeAll 关闭数据库连接
func (m *Manager) closeAll() error {
	var err error
	m.Range(func(name, db any) bool {
		err = db.(*DB).invoker.(io.Closer).Close()
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
	dsn := item.dsn()
	database, err := sql.Open(item.driver(), dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(item.maxIdleConns())
	database.SetMaxOpenConns(item.maxOpenConns())
	if item.connMaxLifetime() > 0 {
		database.SetConnMaxLifetime(time.Duration(item.connMaxLifetime()) * time.Second)
	}
	return database, nil
}

// Init 初始化数据库
func Init[T MySQLConfig | SQLiteConfig](conf []T) error {
	var configs []Config
	for _, item := range conf {
		configs = append(configs, Config(item))
	}
	return manager.init(configs)
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
