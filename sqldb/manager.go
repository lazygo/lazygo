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
func (m *Manager) init[T Config](conf []T) error {
	for _, item := range conf {
		if _, ok := manager.Load(item.GetName()); ok {
			// 已连接的就不再次连接了
			continue
		}
		db, err := manager.open(item)
		if err != nil {
			return err
		}
		err = db.Ping()
		if err != nil {
			db.Close()
			return err
		}
		m.Store(item.GetName(), &DB{Tx{
			invoker: db,
			name:    item.GetName(),
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
	dsn := item.GetDSN()
	database, err := sql.Open(item.GetDriver(), dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxIdleConns(item.GetMaxIdleConns())
	database.SetMaxOpenConns(item.GetMaxOpenConns())
	if item.GetConnMaxLifetime() > 0 {
		database.SetConnMaxLifetime(time.Duration(item.GetConnMaxLifetime()) * time.Second)
	}
	return database, nil
}

// Init 初始化数据库
func Init[T Config](conf []T) error {
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
