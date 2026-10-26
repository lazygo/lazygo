package sqldb

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrNoData = errors.New("not exists")
)

// 查询构建器
type Builder struct {
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
func newBuilder(tx *Tx, table Table) *Builder {
	return &Builder{
		tx:    tx,
		table: table,
		cond:  newGroup(AND),
	}
}

// withBefore 返回携带hook的执行用Tx
// hook非nil时拷贝tx并覆盖其before，避免直接改写共享的Tx（并发竞争与hook残留）
func withBefore(tx *Tx, hook func(string, ...any) func()) *Tx {
	if hook == nil || tx == nil {
		return tx
	}
	ntx := *tx
	ntx.before = hook
	return &ntx
}

// Where 自动识别查询 推荐使用
// cond string                                字符串查询条件
// cond map[string]interface{}                map查询条件，同WhereMap
// field string, value []any                  In查询条件，同WhereIn
// field string, op string value interface{}  同WhereIn
func (b *Builder) Where(cond ...any) *Builder {
	b.lastError = b.cond.where(cond...)
	return b
}

// WhereMap Map查询
// key中不包含运算符时，会自动将map拼接为`k1`='v2' AND `k2`='v2' 的形式
// key中包含条件运算符时，例如 Map{"key >=": 1}  会拼接为 `k2` >= 'v2'
// key中的运算符应与key名之间使用空格隔开，可用的运算符包括 “>” “>=” “<” “<=” “!=” “in” “not in” “like” “contains”
// “contains” 为模糊包含匹配（生成 LIKE ?，参数自动经 LikePattern 转义包裹为 %kw%）；“like” 原样使用传入的值，需自行处理通配符
// map的某个key对应的值为任意类型切片时，会将此key及其对应的切片转换为IN查询条件
func (b *Builder) WhereMap(cond map[string]any) *Builder {
	b.cond.whereMap(cond)
	return b
}

// WhereRaw 子句查询
func (b *Builder) WhereRaw(cond string, args ...any) *Builder {
	b.cond.whereRaw(cond, args...)
	return b
}

// WhereIn IN查询
func (b *Builder) WhereIn(k string, in []any) *Builder {
	b.cond.meta(k, "IN", in...)
	return b
}

// WhereNotIn NOT IN查询
func (b *Builder) WhereNotIn(k string, in []any) *Builder {
	b.cond.meta(k, "NOT IN", in...)
	return b
}

// ClearCond 清空当前where
// （每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *Builder) ClearCond() *Builder {
	b.cond.clear()
	return b
}

// Clear 清空当前where和Limit、Offset等内容
func (b *Builder) Clear() *Builder {
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

func (b *Builder) GroupBy(ks ...string) *Builder {
	for _, k := range ks {
		b.groupBy = append(b.groupBy, buildK(k))
	}
	return b
}

func (b *Builder) OrderByRand() *Builder {
	b.orderBy = append(b.orderBy, "RAND()")
	return b
}

func (b *Builder) OrderBy(k string, direct string) *Builder {
	direct = strings.ToUpper(direct)
	if direct != "ASC" && direct != "DESC" {
		// params error
		b.lastError = ErrInvalidArguments
		return b
	}
	b.orderBy = append(b.orderBy, buildK(k)+" "+direct)
	return b
}

func (b *Builder) Offset(offset int64) *Builder {
	b.offset = offset
	return b
}

func (b *Builder) Limit(limit int64) *Builder {
	b.limit = limit
	return b
}

func (b *Builder) Select(fields ...string) *Builder {
	slices.Sort(fields)
	b.fields = fields
	return b
}

// BeforeHook sql执行前hook，仅对下一次查询或更新操作生效
func (b *Builder) BeforeHook(h func(string, ...any) func()) *Builder {
	if h == nil {
		return b
	}
	b.before = h
	return b
}

// buildCond 构建条件
// 把cond用 AND 连接起来
// 仅拼接字符串，不改变查询构建器中暂存的条件，状态由调用方负责清理
func (b *Builder) buildCond() (string, []any) {
	result := b.cond.String()
	args := b.cond.Args()
	return strings.TrimSpace(result), args
}

// QueryString 拼接查询语句字符串
// 仅拼接字符串，不改变查询构建器中暂存的条件和Limit、Offset等内容
// 状态由查询或更新方法（Find/First/One/Update等）在执行后自动清空
func (b *Builder) QueryString() (string, []any, error) {
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

// Count 查询统计个数
func (b *Builder) Count() (int64, error) {
	if b.table == "" {
		return 0, ErrEmptyTableName
	}

	num, err := b.One("COUNT(*) AS _num")
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(num, 10, 64)
}

// Find 查询并返回多条记录
// field string 返回的字段 示例："*"
func (b *Builder) Find[T any]() ([]T, int, error) {
	defer b.Clear()

	queryString, args, err := b.QueryString()
	if err != nil {
		return nil, 0, err
	}

	return withBefore(b.tx, b.before).Find[T](queryString, args...)
}

// First 查询并返回单条记录
// field string 返回的字段 示例："*"
func (b *Builder) First[T any]() (T, int, error) {
	defer b.Clear()

	queryString, args, err := b.Limit(1).QueryString()
	if err != nil {
		var zero T
		return zero, 0, err
	}

	return withBefore(b.tx, b.before).First[T](queryString, args...)
}

// Exists 判读记录是否存在
func (b *Builder) Exists() (bool, error) {
	_, err := b.One("(0) AS _row")
	if err != nil {
		if errors.Is(err, ErrNoData) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// One 查询并返回单个字段
// field string 返回的字段 示例："name"、"count(*) AS count"
func (b *Builder) One(field string) (string, error) {
	defer b.Clear()

	queryString, args, err := b.Limit(1).Select(field).QueryString()
	if err != nil {
		return "", err
	}

	item, n, err := withBefore(b.tx, b.before).First[map[string]string](queryString, args...)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", ErrNoData
	}
	return item[oneColumnName(field)], nil
}

// Insert 单条插入
// set map[string]interface{} 插入的数据
// 返回插入的id，错误信息
func (b *Builder) Insert(set map[string]any) (int64, error) {
	defer b.Clear()

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
	res, err := withBefore(b.tx, b.before).Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取最后插入的主键id
	id, err := res.LastInsertId()
	if err != nil && strings.Contains(err.Error(), "not supported") {
		return 0, nil
	}
	return id, err
}

// Update 更新
// set map[string]interface{} 更新的字段
// limit （可选参数）限制更新limit
// 返回影响的条数，错误信息
func (b *Builder) Update(set map[string]any, limit ...int) (int64, error) {
	defer b.Clear()

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
	res, err := withBefore(b.tx, b.before).Exec(queryString, args...)
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
func (b *Builder) UpdateRaw(set string, limit ...int) (int64, error) {
	defer b.Clear()

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
	res, err := withBefore(b.tx, b.before).Exec(queryString, args...)
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
func (b *Builder) Increment(column string, amount int64, set ...map[string]any) (int64, error) {
	defer b.Clear()

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
	res, err := withBefore(b.tx, b.before).Exec(queryString, args...)
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
func (b *Builder) Decrement(column string, amount int64, set ...map[string]any) (int64, error) {
	return b.Increment(column, -amount, set...)
}

// Delete 删除
func (b *Builder) Delete(limit ...int) (int64, error) {
	defer b.Clear()

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
	res, err := withBefore(b.tx, b.before).Exec(queryString, args...)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}
