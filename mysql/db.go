package mysql

import (
	"database/sql"
	"errors"
)

type Db struct {
	name string
	db   *sql.DB
}

// 分页返回数据 - 表字段名定义
type ResultRow struct {
}

// 分页返回数据 - 返回结果定义
type ResultData struct {
	List      []map[string]interface{}
	Count     int64
	PerPage   int
	Page      int
	PageCount int
	Start     int
	Mark      int
}

var LostConnection = []string{
	"server has gone away",
	"no connection to the server",
	"Lost connection",
	"is dead or not enabled",
	"Error while sending",
	"decryption failed or bad record mac",
	"server closed the connection unexpectedly",
	"SSL connection has been closed unexpectedly",
	"Error writing data to the connection",
	"Resource deadlock avoided",
	"Transaction() on null",
	"child connection forced to terminate due to client_idle_limit",
	"query_wait_timeout",
	"reset by peer",
	"Physical connection is not usable",
	"TCP Provider: Error code 0x68",
	"ORA-03114",
	"Packets out of order. Expected",
	"Adaptive Server connection failed",
	"Communication link failure",
	"connection refused",
}

func NewDb(name string, db *sql.DB) *Db {
	return &Db{
		name: name,
		db:   db,
	}
}

// 获取查询构建器
func (d *Db) Table(table string) *Builder {
	return NewBuilder(d, table)
}

// 获取查询构建器
func (d *Db) Query(query string) (*sql.Rows, error) {
	var err error
	var rows *sql.Rows
	for retry := 2; retry > 0; retry-- {
		rows, err = d.db.Query(query)
		if err == nil {
			return rows, nil
		}
		if !ContainInArray(err.Error(), LostConnection) {
			break
		}
	}
	if rows != nil {
		rows.Close()
	}
	return nil, err
}

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

func (d *Db) GetRow(query string) (map[string]interface{}, error) {
	rows, err := d.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	outArr, err := parseRowData(rows)

	return outArr, err
}

// 解析结果集
func parseData(rows *sql.Rows) ([]map[string]interface{}, error) {
	var data []map[string]interface{} = nil

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
