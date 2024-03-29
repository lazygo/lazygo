package mysql

import (
	"database/sql"
	"errors"
	"fmt"
)

type DB struct {
	*sql.DB
	name   string // 数据库名称
	prefix string // 表前缀
	slow   int    // 慢查询时间
}

// ResultData 分页返回数据 - 返回结果定义
type ResultData struct {
	List      []map[string]interface{} `json:"list"`
	Count     int64                    `json:"count"`
	PageSize  int64                    `json:"page_size"`
	Page      int64                    `json:"page"`
	PageCount int64                    `json:"page_count"`
	Start     int64                    `json:"start"`
	Mark      int64                    `json:"mark"`
}

// ToMap 分页结果集转化为map
func (r *ResultData) ToMap() map[string]interface{} {
	if r == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"list":       r.List,
		"count":      r.Count,
		"page_size":  r.PageSize,
		"page":       r.Page,
		"page_count": r.PageCount,
		"start":      r.Start,
		"mark":       r.Mark,
	}
}

func newDb(name string, db *sql.DB, prefix string) *DB {
	return &DB{
		DB:     db,
		name:   name,
		prefix: prefix,
		slow:   200,
	}
}

// Table 获取查询构建器
func (d *DB) Table(table string) *builder {
	return newBuilder(d, d.prefix+table)
}

// TableRaw 获取查询构建器
func (d *DB) TableRaw(table string) *builder {
	return newBuilder(d, table)
}

// Query 查询sql并返回结果集
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// start := time.Now()
	// defer func() {
	// 	// 计算查询执行时间，记录慢查询
	// 	since := time.Since(start)
	// 	if since > time.Duration(d.slow)*time.Millisecond {
	// 		log.Println("慢查询 " + strconv.FormatInt(int64(since/time.Millisecond), 10) + " SQL:" + query)
	// 	}
	// }()
	return d.DB.Query(query, args...)
}

// Exec 执行sql
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}

// GetAll 直接执行sql原生语句并返回多行
func (d *DB) GetAll(result interface{}, query string, args ...interface{}) (int, error) {
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
func (d *DB) GetRow(result interface{}, query string, args ...interface{}) (int, error) {
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

// GetTablePrefix 获取表前缀
func (d *DB) GetTablePrefix() string {
	return d.prefix
}
