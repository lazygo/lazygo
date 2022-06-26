package mysql

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ReadBuilder interface {
	MakeQueryString(fields []interface{}) (string, []interface{}, error)
	Count() (int64, error)
	Fetch(fields []interface{}, result interface{}) (int, error)
	FetchRow(fields []interface{}, result interface{}) (int, error)
	FetchOne(field string) (string, error)
	FetchWithPage(fields []interface{}, page int64, pageSize int64) (*ResultData, error)
}

type WriteBuilder interface {
	Update(set map[string]interface{}, limit ...int) (int64, error)
	UpdateRaw(set string, limit ...int) (int64, error)
	Increment(column string, amount int64, set ...map[string]interface{}) (int64, error)
	Decrement(column string, amount int64, set ...map[string]interface{}) (int64, error)
	Delete(limit ...int) (int64, error)
}

type Builder interface {
	Where(cond ...interface{}) Builder
	WhereMap(cond map[string]interface{}) Builder
	WhereRaw(cond string, args ...interface{}) Builder
	WhereIn(k string, in []interface{}) Builder
	WhereNotIn(k string, in []interface{}) Builder
	ClearCond() Builder
	GroupBy(k string, ks ...string) ReadBuilder
	OrderBy(k string, direct string) ReadBuilder
	Offset(offset int64) ReadBuilder
	Limit(limit int64) Builder
	ReadBuilder
	WriteBuilder
}

// 查询构建器
type builder struct {
	handler   *DB
	table     string
	cond      []string      // 查询构建器中暂存的条件，用于链式调用。每调用一次Where，此数组追加元素。调用查询或更新方法后，此条件自动清空
	args      []interface{} // 预处理参数
	orderBy   []string
	groupBy   []string
	offset    int64
	limit     int64
	lastError error
}

type Raw string

// newBuilder 实例化查询构建器
func newBuilder(handler *DB, table string) *builder {
	return &builder{
		handler: handler,
		table:   table,
		cond:    []string{},
	}
}

// Where 自动识别查询 推荐使用
// cond string 字符串查询条件
// cond map[string]interface{} map查询条件，同WhereMap
// cond string, []interface{} In查询条件，同WhereIn
// cond string, []anytype In查询条件，同WhereIn
// cond field string, op string value interface{}，同WhereIn
func (b *builder) Where(cond ...interface{}) Builder {
	switch len(cond) {
	case 1:
		switch cond[0].(type) {
		case string, Raw:
			// 字符串查询
			return b.WhereRaw(cond[0].(string))
		case map[string]interface{}:
			// map拼接查询
			return b.WhereMap(cond[0].(map[string]interface{}))
		default:
			b.lastError = ErrInvalidCondArguments
			return b
		}
	case 2:
		k, ok := cond[0].(string)
		if !ok {
			b.lastError = ErrInvalidCondArguments
			break
		}
		// in查询
		in, ok := CreateAnyTypeSlice(cond[1])
		if ok {
			return b.WhereIn(k, in)
		}
		// k = v
		return b.Where(k, "=", cond[1])
	case 3:
		k, ok := cond[0].(string)
		if !ok {
			break
		}
		op, ok := cond[1].(string)
		if !ok {
			break
		}
		// k op v
		return b.WhereRaw(build(k, op), cond[2])
	default:
	}
	b.lastError = ErrInvalidCondArguments
	return b
}

// WhereMap Map查询
// 会自动将map拼接为`k1`='v2' AND `k2`='v2' 的形式
// map的某个key对应的值为任意类型切片时，会将此key及其对应的切片转换为IN查询条件
func (b *builder) WhereMap(cond map[string]interface{}) Builder {
	for k, v := range cond {
		if vv, ok := CreateAnyTypeSlice(v); ok {
			b.WhereIn(k, vv)
		} else {
			b.cond = append(b.cond, build(k, "="))
			b.args = append(b.args, v)
		}
	}
	return b
}

// WhereRaw 子句查询
func (b *builder) WhereRaw(cond string, args ...interface{}) Builder {
	cond = strings.Trim(cond, " ")
	if cond != "" {
		b.cond = append(b.cond, cond)
		b.args = append(b.args, args...)
	}
	return b
}

// WhereIn IN查询
func (b *builder) WhereIn(k string, in []interface{}) Builder {
	var arr []string
	for range in {
		arr = append(arr, "?")
	}
	cond := fmt.Sprintf("%s IN(%s)", buildK(k), strings.Join(arr, ", "))
	b.cond = append(b.cond, cond)
	b.args = append(b.args, in...)
	return b
}

// WhereNotIn NOT IN查询
func (b *builder) WhereNotIn(k string, in []interface{}) Builder {
	var arr []string
	for range in {
		arr = append(arr, "?")
	}
	cond := fmt.Sprintf("%s NOT IN(%s)", buildK(k), strings.Join(arr, ", "))
	b.cond = append(b.cond, cond)
	b.args = append(b.args, in...)
	return b
}

// ClearCond 清空当前where
//（每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *builder) ClearCond() Builder {
	b.cond = []string{}
	b.args = []interface{}{}
	return b
}

// Clear 清空当前where和Limit、Offset等内容
func (b *builder) Clear() Builder {
	b.ClearCond()
	b.groupBy = []string{}
	b.orderBy = []string{}
	b.limit = 0
	b.offset = 0
	b.lastError = nil
	return b
}

func (b *builder) GroupBy(k string, ks ...string) ReadBuilder {
	b.groupBy = append(b.groupBy, buildK(k))
	for _, k := range ks {
		b.groupBy = append(b.groupBy, buildK(k))
	}
	return b
}
func (b *builder) OrderBy(k string, direct string) ReadBuilder {
	direct = strings.ToUpper(direct)
	if direct != "ASC" && direct != "DESC" {
		// params error
		b.lastError = ErrInvalidArguments
		return b
	}
	b.orderBy = append(b.orderBy, buildK(k)+" "+direct)
	return b
}

func (b *builder) Offset(offset int64) ReadBuilder {
	b.offset = offset
	return b
}

func (b *builder) Limit(limit int64) Builder {
	b.limit = limit
	return b
}

// buildCond 构建条件
// 把cond用 AND 连接起来
// buildCond调用之后会清空当期查询构建器中暂存的条件
//（每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *builder) buildCond() (string, []interface{}) {
	defer b.ClearCond()
	if len(b.cond) == 0 {
		return "", nil
	}
	result := strings.Join(b.cond, " AND ")
	args := b.args
	return strings.TrimSpace(result), args
}

// buildVal 构建值
// val map[string]interface{} 将会被展开成 k=? 的形式 多个元素用逗号隔开
// extra []string 多个元素间用逗号隔开，并追加在 val参数展开的字符串后面
// 示例：val = {"last_view_time": "12345678", "last_view_user": "li"}; extra = ["view_num=view_num+1"]
func buildVal(val map[string]interface{}, extra []string) (string, []interface{}) {
	var items []string
	var args []interface{}

	// 使用等号连接map的key和value，并放入数组
	for k, v := range val {
		items = append(items, build(k, "="))
		args = append(args, v)
	}

	// 追加额外的数组元素
	for _, v := range extra {
		v = strings.Trim(v, " ")
		if v != "" {
			items = append(items, v)
		}
	}
	// 用逗号将数组元素拼接成字符串
	result := strings.Join(items, ", ")
	return result, args
}

// build 构造
// 用操作符op把key和value占位符连接起来
// 示例 k = "name", op = "="    `name`=?
func build(k string, op string) string {
	return buildK(k) + " " + op + " ?"
}

// buildK key增加反引号
func buildK(k string) string {
	k = strings.TrimSpace(k)
	if k == "*" {
		return "*"
	}
	k = strings.ReplaceAll(k, "`", "")
	t := ""
	idx := strings.Index(k, ".")
	if idx != -1 {
		t = k[:idx+1]
		k = k[idx+1:]
	}
	return t + "`" + k + "`"
}

// buildFields 构造查询fields
func buildFields(fields []interface{}) (string, error) {
	arr := make([]string, len(fields), cap(fields))
	for i, v := range fields {
		switch v.(type) {
		case Raw:
			arr[i] = string(v.(Raw))
		case string:
			arr[i] = buildK(v.(string))
		default:
			return "", ErrInvalidColumnsArguments
		}
	}
	return strings.Join(arr, ", "), nil
}

// MakeQueryString 拼接查询语句字符串
func (b *builder) MakeQueryString(fields []interface{}) (string, []interface{}, error) {
	defer b.Clear()

	if b.table == "" {
		return "", nil, ErrEmptyTableName
	}

	if b.lastError != nil {
		return "", nil, b.lastError
	}

	fieldsStr, err := buildFields(fields)
	if err != nil {
		return "", nil, err
	}

	queryString := fmt.Sprintf("SELECT %s FROM %s", fieldsStr, b.table)

	// 构建where条件
	cond, args := b.buildCond()
	if cond != "" {
		queryString += " WHERE " + cond
	}

	if len(b.groupBy) > 0 {
		queryString += " GROUP BY " + strings.Join(b.groupBy, ", ")
	}

	if len(b.orderBy) > 0 {
		queryString += " ORDER BY " + strings.Join(b.orderBy, ", ")
	}

	if b.limit > 0 && b.offset > 0 {
		queryString += fmt.Sprintf(" LIMIT %d, %d", b.offset, b.limit)
	} else if b.limit > 0 {
		queryString += fmt.Sprintf(" LIMIT %d", b.limit)
	}
	return queryString, args, nil
}

// Multi 构造分页返回结果
// count int64 总数
// perpage int64 每页数量
// page int64 当前页
func Multi(count int64, page int64, pageSize int64) ResultData {
	var data ResultData
	data.Count = int64(math.Max(float64(count), 0))                                 // 总数
	data.Page = int64(math.Max(float64(page), 1))                                   // 当前页
	data.PageSize = int64(math.Max(float64(pageSize), 1))                           // 每页数量
	data.PageCount = int64(math.Ceil(float64(data.Count) / float64(data.PageSize))) // 总页数
	data.Start = int64(math.Max(float64(data.PageSize*data.Page-data.PageSize), 0)) // 当前页之前有多少条数据
	data.Mark = data.Start + 1                                                      // 当前页开始是第几条数据
	return data
}

// Count 查询统计个数
func (b *builder) Count() (int64, error) {
	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	result := &struct {
		Num int64 `json:"num"`
	}{}

	_, err := b.FetchRow([]interface{}{Raw("COUNT(*) AS num")}, result)
	if err != nil {
		return 0, err
	}

	return result.Num, nil
}

// Fetch 查询并返回多条记录
// field string 返回的字段 示例："*"
func (b *builder) Fetch(fields []interface{}, result interface{}) (int, error) {

	queryString, args, err := b.MakeQueryString(fields)
	if err != nil {
		return 0, err
	}

	return b.handler.GetAll(result, queryString, args...)
}

// FetchWithPage 查询并返回多条记录，且包含分页信息
// page 第几页，从1开始
// limit 结果数量
func (b *builder) FetchWithPage(fields []interface{}, page int64, pageSize int64) (*ResultData, error) {

	cond, args := b.buildCond()
	count, err := b.ClearCond().WhereRaw(cond, args...).Count()
	if err != nil {
		return nil, err
	}

	data := Multi(count, page, pageSize)
	_, err = b.ClearCond().WhereRaw(cond, args...).Limit(data.PageSize).Offset(data.Start).Fetch(fields, &data.List)
	return &data, err
}

// FetchRow 查询并返回单条记录
// field string 返回的字段 示例："*"
func (b *builder) FetchRow(fields []interface{}, result interface{}) (int, error) {

	queryString, args, err := b.MakeQueryString(fields)
	if err != nil {
		return 0, err
	}

	return b.handler.GetRow(result, queryString, args...)
}

// FetchOne 查询并返回单个字段
// field string 返回的字段 示例："count(*) AS count"
func (b *builder) FetchOne(field string) (string, error) {

	queryString, args, err := b.MakeQueryString([]interface{}{field})
	if err != nil {
		return "", err
	}

	item := map[string]string{}
	_, err = b.handler.GetRow(&item, queryString, args...)
	if err != nil {
		return "", err
	}
	return item[field], nil
}

// Insert 单条插入
// set map[string]interface{} 插入的数据
// 返回插入的id，错误信息
func (b *builder) Insert(set map[string]interface{}) (int64, error) {
	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	// 拼接查询语句
	var fields []string
	var values []string
	var args []interface{}
	for name, value := range set {
		fields = append(fields, buildK(name))
		values = append(values, "?")
		args = append(args, value)
	}
	queryString := "INSERT INTO " + b.table + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(values, ", ") + ")"

	// 执行插入语句
	res, err := b.handler.Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取最后插入的主键id
	id, err := res.LastInsertId()
	return id, err
}

// Update 更新
// set map[string]interface{} 更新的字段
// limit （可选参数）限制更新limit
// 返回影响的条数，错误信息
func (b *builder) Update(set map[string]interface{}, limit ...int) (int64, error) {

	// 构建where条件，先构建where条件，防止执行不到ClearCond
	where, args := b.buildCond()
	if b.lastError != nil {
		return 0, b.lastError
	}
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	if len(limit) > 1 {
		return 0, ErrInvalidArguments
	}

	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	strset, valArgs := buildVal(set, []string{})
	args = append(valArgs, args...)

	// 查询字符串
	queryString := "UPDATE " + b.table + " SET " + strset + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 执行更新语句
	res, err := b.handler.Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}

// UpdateRaw 更新
// set map[string]interface{} 更新的字段
// limit （可选参数）限制更新limit
// 返回影响的条数，错误信息
func (b *builder) UpdateRaw(set string, limit ...int) (int64, error) {
	// 构建where条件，先构建where条件，防止执行不到ClearCond
	where, args := b.buildCond()

	if b.lastError != nil {
		return 0, b.lastError
	}

	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	if len(limit) > 1 {
		return 0, ErrInvalidArguments
	}

	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	// 查询字符串
	queryString := "UPDATE " + b.table + " SET " + set + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 执行更新语句
	res, err := b.handler.Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}

// Increment 自增
// column string 自增的字段
// amount int 自增的数量
// set map[string]interface{} （可选参数）自增同时update的字段
// 返回影响的条数，错误信息
func (b *builder) Increment(column string, amount int64, set ...map[string]interface{}) (int64, error) {
	// 构建where条件，先构建where条件，防止执行不到ClearCond
	where, args := b.buildCond()
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	if len(set) > 1 {
		return 0, ErrInvalidArguments
	}
	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if b.lastError != nil {
		return 0, b.lastError
	}

	column = buildK(column)
	// 拼接自增语句
	var extra []string
	if amount >= 0 {
		extra = []string{
			fmt.Sprintf("%s = %s + %d", column, column, amount),
		}
	} else {
		extra = []string{
			fmt.Sprintf("%s = %s - %d", column, column, -amount),
		}
	}

	// 拼接sql
	strset, valArgs := buildVal(mergeMap(set...), extra)

	var queryString = ""
	if len(set) > 0 && set[0] != nil {
		queryString = "UPDATE " + b.table + " SET " + strset + " WHERE " + where
	} else {
		queryString = "UPDATE " + b.table + " SET " + extra[0] + " WHERE " + where
	}

	// 执行更新sql语句
	args = append(valArgs, args...)
	res, err := b.handler.Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}

// Decrement 自减
// column string 自减的字段
// amount int 自减的数量
// set map[string]interface{} （可选参数）自减同时update的字段
// 返回影响的条数，错误信息
func (b *builder) Decrement(column string, amount int64, set ...map[string]interface{}) (int64, error) {
	return b.Increment(column, -amount, set...)
}

// Delete 删除
func (b *builder) Delete(limit ...int) (int64, error) {
	// 构建where条件，先构建where条件，防止执行不到ClearCond
	where, args := b.buildCond()
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	if len(limit) > 1 {
		return 0, ErrInvalidArguments
	}

	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if b.lastError != nil {
		return 0, b.lastError
	}

	// 拼接删除语句
	queryString := "DELETE FROM " + b.table + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 获取影响的行数
	res, err := b.handler.Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}
