package sqldb

import (
	"cmp"
	"fmt"
)

type MySQLConfig struct {
	Driver          string            `json:"driver" toml:"driver"`
	Name            string            `json:"name" toml:"name"`
	User            string            `json:"user" toml:"user"`
	Passwd          string            `json:"passwd" toml:"passwd"`
	Schema          string            `json:"schema" toml:"schema"`
	Host            string            `json:"host" toml:"host"`
	Port            int               `json:"port" toml:"port"`
	DbName          string            `json:"dbname" toml:"dbname"`
	Params          map[string]string `json:"params" toml:"params"`
	MaxOpenConns    int               `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int               `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime int               `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

func (c MySQLConfig) GetName() string {
	return c.Name
}
func (c MySQLConfig) GetDriver() string {
	return c.Driver
}

func (c MySQLConfig) GetDSN() string {
	dsn := fmt.Sprintf(
		"%s:%s@%s(%s:%d)/%s",
		c.User,
		c.Passwd,
		cmp.Or(c.Schema, "tcp"),
		c.Host,
		c.Port,
		c.DbName,
	)
	if len(c.Params) > 0 {
		dsn += "?" + buildQuery(c.Params)
	}
	return dsn
}

func (c MySQLConfig) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

func (c MySQLConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

func (c MySQLConfig) GetConnMaxLifetime() int {
	return c.ConnMaxLifetime
}

type SQLiteConfig struct {
	Driver          string            `json:"driver" toml:"driver"`
	Name            string            `json:"name" toml:"name"`
	Path            string            `json:"path" toml:"path"`
	Params          map[string]string `json:"params" toml:"params"`
	MaxOpenConns    int               `json:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int               `json:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime int               `json:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

func (c SQLiteConfig) GetName() string {
	return c.Name
}

func (c SQLiteConfig) GetDriver() string {
	return c.Driver
}

func (c SQLiteConfig) GetDSN() string {
	dsn := c.Path
	if len(c.Params) > 0 {
		dsn += "?" + buildQuery(c.Params)
	}
	return dsn
}

func (c SQLiteConfig) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

func (c SQLiteConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

func (c SQLiteConfig) GetConnMaxLifetime() int {
	return c.ConnMaxLifetime
}

type Config interface {
	GetName() string
	GetDriver() string
	GetDSN() string
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() int
}
