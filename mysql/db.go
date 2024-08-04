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

// ResultData 分页返回数据 - 返回结果定义
type ResultData struct {
	List      []map[string]any `json:"list"`
	Count     int64            `json:"count"`
	PageSize  int64            `json:"page_size"`
	Page      int64            `json:"page"`
	PageCount int64            `json:"page_count"`
	Start     int64            `json:"start"`
	Mark      int64            `json:"mark"`
}

// ToMap 分页结果集转化为map
func (r *ResultData) ToMap() map[string]any {
	if r == nil {
		return map[string]any{}
	}
	return map[string]any{
		"list":       r.List,
		"count":      r.Count,
		"page_size":  r.PageSize,
		"page":       r.Page,
		"page_count": r.PageCount,
		"start":      r.Start,
		"mark":       r.Mark,
	}
}

// Table 获取查询构建器
func (d *DB) Table(table string) *builder {
	return newBuilder(d, Table(table))
}

// Before sql执行前hook
func (d *DB) Before(h func(string, ...any) func()) {
	if h == nil {
		return
	}
	d.before = h
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

// GetAll 直接执行sql原生语句并返回多行
func (d *DB) GetAll(result any, query string, args ...any) (int, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = rows.Close()
	}()
	return parseData(rows, result)
}

// GetRow 直接执行sql原生语句并返回1行
func (d *DB) GetRow(result any, query string, args ...any) (int, error) {
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
func (d *DB) Transaction(fn func() error) (err error) {
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
	err = fn()
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return rbErr
		}
	}
	return tx.Commit()
}
