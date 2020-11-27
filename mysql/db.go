package mysql

import (
	"database/sql"
	"errors"
	"github.com/lazygo/lazygo/utils"
	"strconv"
	"time"
)

type Db struct {
	name   string // 数据库名称
	db     *sql.DB
	prefix string // 表前缀
	slow   int    // 慢查询时间
}

// 分页返回数据 - 表字段名定义
type ResultRow struct {
}

// 分页返回数据 - 返回结果定义
type ResultData struct {
	List      []map[string]interface{} `json:"list"`
	Count     int64                    `json:"count"`
	PerPage   int                      `json:"pre_page"`
	Page      int                      `json:"page"`
	PageCount int                      `json:"page_count"`
	Start     int                      `json:"start"`
	Mark      int                      `json:"mark"`
}

// 分页结果集转化为map
func (r *ResultData) ToMap() map[string]interface{} {
	if r == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"list":       r.List,
		"count":      r.Count,
		"pre_page":   r.PerPage,
		"page":       r.Page,
		"page_count": r.PageCount,
		"start":      r.Start,
		"mark":       r.Mark,
	}
}

func newDb(name string, db *sql.DB, prefix string) *Db {
	return &Db{
		name:   name,
		db:     db,
		prefix: prefix,
		slow:   200,
	}
}

// 获取查询构建器
func (d *Db) Table(table string) *Builder {
	return newBuilder(d, d.prefix+table)
}

// 查询sql并返回结果集
func (d *Db) Query(query string) (*sql.Rows, error) {
	start := time.Now()
	defer func() {
		// 计算查询执行时间，记录慢查询
		since := time.Since(start)
		if since > time.Duration(d.slow)*time.Millisecond {
			utils.Warn("慢查询 " + strconv.FormatInt(int64(since/time.Millisecond), 10) + " SQL:" + query)
		}
	}()
	return d.db.Query(query)
}

// 执行sql
func (d *Db) Exec(query string) (sql.Result, error) {
	stmt, err := d.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 直接执行sql原生语句并返回多行
func (d *Db) GetAll(query string) ([]map[string]interface{}, error) {
	rows, err := d.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	outArr, err := parseData(rows)
	if err != nil {
		return nil, err
	}

	return outArr, nil
}

// 直接执行sql原生语句并返回1行
func (d *Db) GetRow(query string) (map[string]interface{}, error) {
	rows, err := d.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	outArr, err := parseRowData(rows)

	return outArr, err
}

// GetTablePrefix 获取表前缀
func (d *Db) GetTablePrefix() string {
	return d.prefix
}

// 解析结果集
func parseData(rows *sql.Rows) ([]map[string]interface{}, error) {
	data := make([]map[string]interface{}, 0, 20)

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	fCount := len(columns)
	fieldPtr := make([]interface{}, fCount)
	fieldArr := make([]sql.RawBytes, fCount)
	fieldToID := make(map[string]int64, fCount)
	for k, v := range columns {
		fieldPtr[k] = &fieldArr[k]
		fieldToID[v] = int64(k)
	}

	for rows.Next() {
		err = rows.Scan(fieldPtr...)
		if err != nil {
			return nil, err
		}

		m := make(map[string]interface{}, fCount)

		for k, v := range fieldToID {
			if fieldArr[v] == nil {
				m[k] = ""
			} else {
				m[k] = string(fieldArr[v])
			}
		}
		data = append(data, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// 解析单行结果集
func parseRowData(rows *sql.Rows) (map[string]interface{}, error) {
	if rows == nil {
		return nil, errors.New("数据库初始化失败")
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	fCount := len(columns)
	fieldPtr := make([]interface{}, fCount)
	fieldArr := make([]sql.RawBytes, fCount)
	fieldToID := make(map[string]int64, fCount)
	for k, v := range columns {
		fieldPtr[k] = &fieldArr[k]
		fieldToID[v] = int64(k)
	}

	m := make(map[string]interface{}, fCount)
	if rows.Next() {
		err = rows.Scan(fieldPtr...)
		if err != nil {
			return nil, err
		}

		for k, v := range fieldToID {
			if fieldArr[v] == nil {
				m[k] = ""
			} else {
				m[k] = string(fieldArr[v])
			}
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return m, nil
}
