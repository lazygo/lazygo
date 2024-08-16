package mysql

import (
	"database/sql"
	"errors"
	"fmt"
)

type DB struct {
	*sql.DB
	name   string // 数据库名称
	before func(string, ...any) func()
}

// Table 获取查询构建器
func (d *DB) Table(table string) *builder {
	return newBuilder(d, Table(table))
}

// Before sql执行前hook
func (d *DB) Before(h func(string, ...any) func()) *DB {
	if h == nil {
		return d
	}
	d.before = h
	return d
}

// Query 查询sql并返回结果集
func (d *DB) Query(sql string, args ...any) (*sql.Rows, error) {
	after := d.before(sql, args...)
	row, err := d.DB.Query(sql, args...)
	if after != nil {
		after()
	}
	if err != nil {
		return nil, &SqlError{
			err:  err,
			sql:  sql,
			args: args,
		}
	}
	return row, nil
}

// Exec 执行sql
func (d *DB) Exec(sql string, args ...any) (sql.Result, error) {
	after := d.before(sql, args...)
	result, err := d.DB.Exec(sql, args...)
	if after != nil {
		after()
	}
	if err != nil {
		return nil, &SqlError{
			err:  err,
			sql:  sql,
			args: args,
		}
	}
	return result, nil
}

// Find 直接执行sql原生语句并返回多行
func (d *DB) Find(result any, query string, args ...any) (int, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rows.Close()
	}()
	return parseData(rows, result)
}

// First 直接执行sql原生语句并返回1行
func (d *DB) First(result any, query string, args ...any) (int, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rows.Close()
	}()
	return parseRowData(rows, result)
}

// Transaction 事务
func (d *DB) Transaction(fn func(db *DB) error) (err error) {
	tx, err := d.Begin()
	if err != nil {
		if tx != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return rbErr
			}
		}
		return err
	}
	defer func() {
		e := recover()
		if e != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = rbErr
				return
			}
			err = errors.New(fmt.Sprint(e))
		}
	}()
	err = fn(d)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return rbErr
		}
	}
	return tx.Commit()
}
