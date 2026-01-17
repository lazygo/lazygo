package mysql

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

// RBuilder read builder (R)
type RBuilder interface {
	QueryString() (string, []any, error)
	Count() (int64, error)
	Find(result any) (int, error)
	First(result any) (int, error)
	One(field string) (string, error)
	FetchWithPage(page int64, pageSize int64) (*ResultData, error)
}

// UDBuilder update or delete builder (UD)
type UDBuilder interface {
	Update(set map[string]any, limit ...int) (int64, error)
	UpdateRaw(set string, limit ...int) (int64, error)
	Increment(column string, amount int64, set ...map[string]any) (int64, error)
	Decrement(column string, amount int64, set ...map[string]any) (int64, error)
	Delete(limit ...int) (int64, error)
}

// CBuilder create builder (C)
type CBuilder interface {
	Insert(set map[string]any) (int64, error)
}

type WhereBuilder interface {
	Where(cond ...any) WhereBuilder
	WhereMap(cond map[string]any) WhereBuilder
	WhereRaw(cond string, args ...any) WhereBuilder
	WhereIn(k string, in []any) WhereBuilder
	WhereNotIn(k string, in []any) WhereBuilder
	ClearCond() WhereBuilder
	GroupBy(ks ...string) WhereBuilder
	OrderBy(k string, direct string) WhereBuilder
	OrderByRand() WhereBuilder
	Offset(offset int64) WhereBuilder
	Limit(limit int64) WhereBuilder
	Select(fields ...string) WhereBuilder
	RBuilder
	UDBuilder
}

type Builder interface {
	BeforeHook(h func(string, ...any) func()) Builder
	WhereBuilder
	CBuilder
}

// 查询构建器
type builder struct {
	tx        *Tx
	table     Table
	cond      *groupCond // 查询构建器中暂存的条件，用于链式调用。每调用一次Where，此数组追加元素。调用查询或更新方法后，此条件自动清空
	fields    Fields
	orderBy   []string
	groupBy   []string
	offset    int64
	limit     int64
	before    func(string, ...any) func()
	lastError error
}

// newBuilder 实例化查询构建器
func newBuilder(tx *Tx, table Table) *builder {
	return &builder{
		tx:    tx,
		table: table,
		cond:  newGroup(AND),
	}
}

// Where 自动识别查询 推荐使用
// cond string                                字符串查询条件
// cond map[string]interface{}                map查询条件，同WhereMap
// field string, value []any                  In查询条件，同WhereIn
// field string, op string value interface{}  同WhereIn
func (b *builder) Where(cond ...any) WhereBuilder {
	b.lastError = b.cond.where(cond...)
	return b
}

// WhereMap Map查询
// key中不包含运算符时，会自动将map拼接为`k1`='v2' AND `k2`='v2' 的形式
// key中包含条件运算符时，例如 Map{"key >=": 1}  会拼接为 `k2` >= 'v2'
// key中的运算符应与key名之间使用空格隔开，可用的运算符包括 “>” “>=” “<” “<=” “!=” “in” “not in” “like”
// map的某个key对应的值为任意类型切片时，会将此key及其对应的切片转换为IN查询条件
func (b *builder) WhereMap(cond map[string]any) WhereBuilder {
	b.cond.whereMap(cond)
	return b
}

// WhereRaw 子句查询
func (b *builder) WhereRaw(cond string, args ...any) WhereBuilder {
	b.cond.whereRaw(cond, args...)
	return b
}

// WhereIn IN查询
func (b *builder) WhereIn(k string, in []any) WhereBuilder {
	b.cond.meta(k, "IN", in...)
	return b
}

// WhereNotIn NOT IN查询
func (b *builder) WhereNotIn(k string, in []any) WhereBuilder {
	b.cond.meta(k, "NOT IN", in...)
	return b
}

// ClearCond 清空当前where
// （每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *builder) ClearCond() WhereBuilder {
	b.cond.clear()
	return b
}

// Clear 清空当前where和Limit、Offset等内容
func (b *builder) Clear() Builder {
	b.ClearCond()
	b.fields = nil
	b.groupBy = []string{}
	b.orderBy = []string{}
	b.offset = 0
	b.limit = 0
	b.before = nil
	b.lastError = nil
	return b
}

func (b *builder) GroupBy(ks ...string) WhereBuilder {
	for _, k := range ks {
		b.groupBy = append(b.groupBy, buildK(k))
	}
	return b
}

func (b *builder) OrderByRand() WhereBuilder {
	b.orderBy = append(b.orderBy, "RAND()")
	return b
}

func (b *builder) OrderBy(k string, direct string) WhereBuilder {
	direct = strings.ToUpper(direct)
	if direct != "ASC" && direct != "DESC" {
		// params error
		b.lastError = ErrInvalidArguments
		return b
	}
	b.orderBy = append(b.orderBy, buildK(k)+" "+direct)
	return b
}

func (b *builder) Offset(offset int64) WhereBuilder {
	b.offset = offset
	return b
}

func (b *builder) Limit(limit int64) WhereBuilder {
	b.limit = limit
	return b
}

func (b *builder) Select(fields ...string) WhereBuilder {
	slices.Sort(fields)
	b.fields = fields
	return b
}

// BeforeHook sql执行前hook
func (b *builder) BeforeHook(h func(string, ...any) func()) Builder {
	if h == nil {
		return b
	}
	b.before = h
	return b
}

// buildCond 构建条件
// 把cond用 AND 连接起来
// buildCond调用之后会清空当期查询构建器中暂存的条件
// （每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *builder) buildCond() (string, []any) {
	defer b.ClearCond()
	result := b.cond.String()
	args := b.cond.Args()
	return strings.TrimSpace(result), args
}

// QueryString 拼接查询语句字符串
func (b *builder) QueryString() (string, []any, error) {
	defer b.Clear()

	if b.table == "" {
		return "", nil, ErrEmptyTableName
	}

	if b.lastError != nil {
		return "", nil, b.lastError
	}

	queryString := fmt.Sprintf("SELECT %s FROM %s", b.fields, b.table)

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

	_, err := b.Select("COUNT(*) AS num").First(result)
	if err != nil {
		return 0, err
	}

	return result.Num, nil
}

// Find 查询并返回多条记录
// field string 返回的字段 示例："*"
func (b *builder) Find(result any) (int, error) {

	queryString, args, err := b.QueryString()
	if err != nil {
		return 0, err
	}

	return b.tx.Before(b.before).Find(result, queryString, args...)
}

// FetchWithPage 查询并返回多条记录，且包含分页信息
// page 第几页，从1开始
// limit 结果数量
func (b *builder) FetchWithPage(page int64, pageSize int64) (*ResultData, error) {
	cond, args := b.buildCond()
	count, err := b.ClearCond().WhereRaw(cond, args...).Count()
	if err != nil {
		return nil, err
	}

	data := Multi(count, page, pageSize)
	_, err = b.ClearCond().WhereRaw(cond, args...).Limit(data.PageSize).Offset(data.Start).Find(&data.List)
	return &data, err
}

// First 查询并返回单条记录
// field string 返回的字段 示例："*"
func (b *builder) First(result any) (int, error) {
	queryString, args, err := b.Limit(1).QueryString()
	if err != nil {
		return 0, err
	}

	return b.tx.Before(b.before).First(result, queryString, args...)
}

// One 查询并返回单个字段
// field string 返回的字段 示例："count(*) AS count"
func (b *builder) One(field string) (string, error) {
	queryString, args, err := b.Limit(1).Select(field).QueryString()
	if err != nil {
		return "", err
	}

	item := map[string]string{}
	_, err = b.tx.Before(b.before).First(&item, queryString, args...)
	if err != nil {
		return "", err
	}
	return item[field], nil
}

// Insert 单条插入
// set map[string]interface{} 插入的数据
// 返回插入的id，错误信息
func (b *builder) Insert(set map[string]any) (int64, error) {
	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	// 拼接查询语句
	var fields []string
	var values []string
	var args []any
	for name, value := range set {
		fields = append(fields, buildK(name))
		values = append(values, "?")
		args = append(args, value)
	}
	queryString := "INSERT INTO " + b.table.String() + " (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(values, ", ") + ")"

	// 执行插入语句
	res, err := b.tx.Before(b.before).Exec(queryString, args...)
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
func (b *builder) Update(set map[string]any, limit ...int) (int64, error) {

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
	queryString := "UPDATE " + b.table.String() + " SET " + strset + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 执行更新语句
	res, err := b.tx.Before(b.before).Exec(queryString, args...)
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
	queryString := "UPDATE " + b.table.String() + " SET " + set + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 执行更新语句
	res, err := b.tx.Before(b.before).Exec(queryString, args...)
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
func (b *builder) Increment(column string, amount int64, set ...map[string]any) (int64, error) {
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
		queryString = "UPDATE " + b.table.String() + " SET " + strset + " WHERE " + where
	} else {
		queryString = "UPDATE " + b.table.String() + " SET " + extra[0] + " WHERE " + where
	}

	// 执行更新sql语句
	args = append(valArgs, args...)
	res, err := b.tx.Before(b.before).Exec(queryString, args...)
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
func (b *builder) Decrement(column string, amount int64, set ...map[string]any) (int64, error) {
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
	queryString := "DELETE FROM " + b.table.String() + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit[0])
	}

	// 获取影响的行数
	res, err := b.tx.Before(b.before).Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}
