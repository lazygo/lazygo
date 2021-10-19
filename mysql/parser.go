package mysql

import (
	"database/sql"
	"reflect"
)

// parseData 解析结果集
func parseData(rows *sql.Rows, result interface{}) (int, error) {
	// row slice pointer value
	rspv := reflect.ValueOf(result)
	if rspv.Kind() != reflect.Ptr || rspv.IsNil() {
		return 0, ErrInvalidResultPtr
	}

	// row slice value
	rsv := rspv.Elem()
	if rsv.Kind() != reflect.Slice {
		return 0, ErrInvalidResultPtr
	}
	rsv.Set(reflect.MakeSlice(rsv.Type(), 0, rsv.Cap()))

	// row type
	rt := rsv.Type().Elem()
	if rt.Kind() == reflect.Struct {
		// row value
		rv := reflect.New(rt).Elem()

		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		fieldPtr, err := getFieldPtr(columns, rv)
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			err = rows.Scan(fieldPtr...)
			if err != nil {
				return 0, err
			}
			rsv.Set(reflect.Append(rsv, rv))
		}
		return rsv.Len(), rows.Err()
	}
	if rt.Kind() == reflect.Map {
		if err := checkMap(rt); err != nil {
			return 0, err
		}
		// row value
		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}

		fieldPtr, fieldArr, fieldToID := getResultPtr(columns)

		for rows.Next() {
			err = rows.Scan(fieldPtr...)
			if err != nil {
				return 0, err
			}

			rv := reflect.MakeMapWithSize(rt, len(columns))
			for k, v := range fieldToID {
				if fieldArr[v] == nil {
					rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(""))
				} else {
					rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(string(fieldArr[v])))
				}
			}

			rsv.Set(reflect.Append(rsv, rv))
		}
		return rsv.Len(), rows.Err()
	}
	return 0, ErrInvalidResultPtr

}

// parseRowData 解析单行结果集
func parseRowData(rows *sql.Rows, result interface{}) (int, error) {
	// result pointer value
	rpv := reflect.ValueOf(result)
	if rpv.Kind() != reflect.Ptr || rpv.IsNil() {
		return 0, ErrInvalidResultPtr
	}

	// result value
	rv := rpv.Elem()
	if rv.Kind() == reflect.Struct {

		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}
		fieldPtr, err := getFieldPtr(columns, rv)
		if err != nil {
			return 0, err
		}
		if !rows.Next() {
			return 0, rows.Err()
		}
		err = rows.Scan(fieldPtr...)
		if err != nil {
			return 0, err
		}
		return 1, rows.Err()
	}
	if rv.Kind() == reflect.Map && !rv.IsNil() {
		if err := checkMap(rv.Type()); err != nil {
			return 0, err
		}
		// row value
		columns, err := rows.Columns()
		if err != nil {
			return 0, err
		}
		fieldPtr, fieldArr, fieldToID := getResultPtr(columns)
		if !rows.Next() {
			return 0, rows.Err()
		}
		err = rows.Scan(fieldPtr...)
		if err != nil {
			return 0, err
		}

		for k, v := range fieldToID {
			if fieldArr[v] == nil {
				rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(""))
			} else {
				rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(string(fieldArr[v])))
			}
		}
		return 1, rows.Err()
	}
	return 0, ErrInvalidResultPtr
}

// checkMap 检查map类型
func checkMap(rt reflect.Type) error {
	if rt.Key().Kind() != reflect.String {
		return ErrInvalidResultPtr
	}
	if rt.Elem().Kind() != reflect.Interface && rt.Elem().Kind() != reflect.String {
		return ErrInvalidResultPtr
	}
	return nil
}

// getResultPtr 解析结果集
func getResultPtr(columns []string) ([]interface{}, []sql.RawBytes, map[string]int64) {
	fCount := len(columns)
	fieldPtr := make([]interface{}, fCount)
	fieldArr := make([]sql.RawBytes, fCount)
	fieldToID := make(map[string]int64, fCount)
	for k, v := range columns {
		fieldPtr[k] = &fieldArr[k]
		fieldToID[v] = int64(k)
	}
	return fieldPtr, fieldArr, fieldToID
}

// getFieldPtr 获取结果集指针数组
func getFieldPtr(columns []string, rv reflect.Value) ([]interface{}, error) {
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
