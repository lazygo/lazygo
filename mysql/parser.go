package mysql

import (
	"database/sql"
	"reflect"
)

// parseData 解析结果集
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

// parseDataIn 解析结果集
func parseDataIn(rows *sql.Rows, result interface{}) error {
	// row slice pointer value
	rspv := reflect.ValueOf(result)
	if rspv.Kind() != reflect.Ptr {
		return ErrInvalidPtr
	}

	// row slice value
	rsv := rspv.Elem()
	if rsv.Kind() != reflect.Slice {
		return ErrInvalidStructSlicePtr
	}

	// row type
	rt := rsv.Type().Elem()
	if rt.Kind() != reflect.Struct {
		return ErrInvalidStructSlicePtr
	}

	// row value
	rv := reflect.New(rt).Elem()

	fieldPtr, err := getFieldPtr(rows, rv)
	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.Scan(fieldPtr...)
		if err != nil {
			return err
		}
		rsv.Set(reflect.Append(rsv, rv))
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

// parseRowData 解析单行结果集
func parseRowData(rows *sql.Rows) (map[string]interface{}, error) {
	if !rows.Next() {
		err := rows.Err()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{}, nil
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
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// parseRowDataIn 解析单行结果集
func parseRowDataIn(rows *sql.Rows, result interface{}) error {
	if !rows.Next() {
		return rows.Err()
	}

	// result pointer value
	rpv := reflect.ValueOf(result)
	if rpv.Kind() != reflect.Ptr {
		return ErrInvalidPtr
	}

	// result value
	rv := rpv.Elem()
	if rv.Kind() != reflect.Struct {
		return ErrInvalidStructPtr
	}

	fieldPtr, err := getFieldPtr(rows, rv)
	if err != nil {
		return err
	}

	err = rows.Scan(fieldPtr...)
	if err != nil {
		return err
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

// getFieldPtr 获取结果集指针数组
func getFieldPtr(rows *sql.Rows, rv reflect.Value) ([]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	fCount := len(columns)

	fieldPtr := make([]interface{}, fCount)

	resultFieldNum := rv.NumField()
	fieldArr := make(map[string]interface{}, resultFieldNum)
	for i := 0; i < resultFieldNum; i++ {
		field := rv.Type().Field(i).Tag.Get("json")
		if field == "" {
			continue
		}
		fieldArr[field] = rv.Field(i).Addr().Interface()
	}

	for k, v := range columns {
		if fv, ok := fieldArr[v]; ok {
			fieldPtr[k] = fv
		} else {
			fieldPtr[k] = &sql.RawBytes{}
		}
	}

	return fieldPtr, nil
}
