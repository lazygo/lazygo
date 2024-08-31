package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type DB struct{ Tx }

type Tx struct {
	invoker
	name   string // 数据库名称
	before func(string, ...any) func()
}

type invoker interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Table 获取查询构建器
func (d *Tx) Table(table string) *builder {
	return newBuilder(d, Table(table))
}

// Before sql执行前hook
func (d *Tx) Before(h func(string, ...any) func()) *Tx {
	if h == nil {
		return d
	}
	d.before = h
	return d
}

// Query 查询sql并返回结果集
func (d *Tx) Query(sql string, args ...any) (*sql.Rows, error) {
	after := d.before(sql, args...)
	row, err := d.invoker.Query(sql, args...)
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
func (d *Tx) Exec(sql string, args ...any) (sql.Result, error) {
	after := d.before(sql, args...)
	result, err := d.invoker.Exec(sql, args...)
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
func (d *Tx) Find(result any, query string, args ...any) (int, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	return parseData(rows, result)
}

// First 直接执行sql原生语句并返回1行
func (d *Tx) First(result any, query string, args ...any) (int, error) {
	rows, err := d.Query(query, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	return parseRowData(rows, result)
}

// Transaction 事务
func (d *DB) Transaction(fn func(tx *Tx) error) (err error) {
	tx, err := d.invoker.(*sql.DB).Begin()
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
				err = fmt.Errorf("panic: %v: %w", e, rbErr)
				return
			}
			err = errors.New(fmt.Sprint(e))
		}
	}()
	err = fn(&Tx{
		invoker: tx,
		name:    d.name,
		before:  d.before,
	})
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}
	return tx.Commit()
}
