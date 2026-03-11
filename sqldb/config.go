package sqldb

import "fmt"

type MySQLConfig struct {
	Name            string `json:"name" toml:"name"`
	User            string `json:"user" toml:"user"`
	Passwd          string `json:"passwd" toml:"passwd"`
	Host            string `json:"host" toml:"host"`
	Port            int    `json:"port" toml:"port"`
	DbName          string `json:"dbname" toml:"dbname"`
	Charset         string `json:"charset" toml:"charset"`
	MaxOpenConns    int    `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

func (c MySQLConfig) name() string {
	return c.Name
}
func (c MySQLConfig) driver() string {
	return "mysql"
}

func (c MySQLConfig) dsn() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=5s",
		c.User,
		c.Passwd,
		c.Host,
		c.Port,
		c.DbName,
		c.Charset,
	)
}

func (c MySQLConfig) maxOpenConns() int {
	return c.MaxOpenConns
}

func (c MySQLConfig) maxIdleConns() int {
	return c.MaxIdleConns
}

func (c MySQLConfig) connMaxLifetime() int {
	return c.ConnMaxLifetime
}

type SQLiteConfig struct {
	Name            string `json:"name" toml:"name"`
	Path            string `json:"path" toml:"path"`
	MaxOpenConns    int    `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

func (c SQLiteConfig) name() string {
	return c.Name
}

func (c SQLiteConfig) driver() string {
	return "sqlite3"
}

func (c SQLiteConfig) dsn() string {
	return c.Path
}

func (c SQLiteConfig) maxOpenConns() int {
	return c.MaxOpenConns
}

func (c SQLiteConfig) maxIdleConns() int {
	return c.MaxIdleConns
}

func (c SQLiteConfig) connMaxLifetime() int {
	return c.ConnMaxLifetime
}

type Config interface {
	name() string
	driver() string
	dsn() string
	maxOpenConns() int
	maxIdleConns() int
	connMaxLifetime() int
}
