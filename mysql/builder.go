package mysql

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// 防止在update、delete操作时，漏掉条件造成的严重后果
// 如果确实不需要条件，请将条件设置为 1=1
var ErrEmptyCond = errors.New("条件不能为空")
var ErrEmptyValue = errors.New("值不能为空")

// 查询构建器
type Builder struct {
	schema *Db
	table  string
	cond   []string // 查询构建器中暂存的条件，用于链式调用。每调用一次Where，此数组追加元素。调用查询或更新方法后，此条件自动清空
}

type Raw string

// newBuilder 实例化查询构建器
func newBuilder(schema *Db, table string) *Builder {
	return &Builder{
		schema: schema,
		table:  table,
		cond:   []string{},
	}
}

// Where 自动识别查询 推荐使用
// cond string 字符串查询条件
// cond map[string]interface{} map查询条件，同WhereMap
// cond string, []interface{} In查询条件，同WhereIn
// cond string, []anytype In查询条件，同WhereIn
func (b *Builder) Where(cond ...interface{}) *Builder {
	switch len(cond) {
	case 1:
		switch cond[0].(type) {
		case string:
			// 字符串查询
			return b.WhereRaw(cond[0].(string))
		case map[string]interface{}:
			// map拼接查询
			return b.WhereMap(cond[0].(map[string]interface{}))
		default:
			panic("invalid arguments")
		}
	case 2:
		k, ok := cond[0].(string)
		if !ok {
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
		v := toString(cond[2])
		// k op v
		return b.WhereRaw(build(k, op, v))
	default:
	}
	panic("invalid arguments")
}

// WhereMap Map查询
// 会自动将map拼接为`k1`='v2' AND `k2`='v2' 的形式
// map的某个key对应的值为任意类型切片时，会将此key及其对应的切片转换为IN查询条件
func (b *Builder) WhereMap(cond map[string]interface{}) *Builder {
	for k, v := range cond {
		if vv, ok := CreateAnyTypeSlice(v); ok {
			b.WhereIn(k, vv)
		} else {
			b.cond = append(b.cond, build(k, "=", v))
		}
	}
	return b
}

// WhereRaw 子句查询
func (b *Builder) WhereRaw(cond string) *Builder {
	cond = strings.Trim(cond, " ")
	if cond != "" {
		b.cond = append(b.cond, cond)
	}
	return b
}

// WhereIn IN查询
func (b *Builder) WhereIn(k string, in []interface{}) *Builder {
	var arr []string
	for _, v := range in {
		arr = append(arr, Addslashes(toString(v)))
	}
	cond := fmt.Sprintf("%s IN('%s')", buildK(k), strings.Join(arr, "', '"))
	b.cond = append(b.cond, cond)
	return b
}

// WhereNotIn NOT IN查询
func (b *Builder) WhereNotIn(k string, in []interface{}) *Builder {
	var arr []string
	for _, v := range in {
		arr = append(arr, Addslashes(toString(v)))
	}
	cond := fmt.Sprintf("%s NOT IN('%s')", buildK(k), strings.Join(arr, "', '"))
	b.cond = append(b.cond, cond)
	return b
}

// Clear 清空当前where
//（每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *Builder) Clear() *Builder {
	b.cond = []string{}
	return b
}

// buildCond 构建条件
// 把cond用 AND 连接起来
// buildCond调用之后会清空当期查询构建器中暂存的条件
//（每次调用Where会向当前查询构建器中暂存条件，用于链式调用）
func (b *Builder) buildCond() string {
	if len(b.cond) == 0 {
		return ""
	}
	result := strings.Join(b.cond, " AND ")
	b.Clear()
	return strings.TrimSpace(result)
}

// buildVal 构建值
// val map[string]interface{} 将会被展开成 k=v 的形式 多个元素用逗号隔开
// extra []string 多个元素间用逗号隔开，并追加在 val参数展开的字符串后面
// 示例：val = {"last_view_time": "12345678", "last_view_user": "li"}; extra = ["view_num=view_num+1"]
func buildVal(val map[string]interface{}, extra []string) string {
	var items []string

	// 使用等号连接map的key和value，并放入数组
	for k, v := range val {
		items = append(items, build(k, "=", v))
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
	return result
}

// build 构造
// 用操作符op把key和value连接起来
// 示例 k = "name", op = "=", v ="li"    `name`='li'
func build(k string, op string, v interface{}) string {
	str := Addslashes(toString(v))
	return fmt.Sprintf("%s %s '%s'", buildK(k), op, str)
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
	return fmt.Sprintf("%s`%s`", t, k)
}

// buildFields 构造查询fields
func buildFields(fields []interface{}) string {
	arr := make([]string, len(fields), cap(fields))
	for i, v := range fields {
		str := toString(v)
		switch v.(type) {
		case Raw:
			arr[i] = str
		default:
			arr[i] = buildK(strings.TrimSpace(str))
		}
	}
	return strings.Join(arr, ", ")
}

// MakeQueryString 拼接查询语句字符串
func (b *Builder) MakeQueryString(fields []interface{}, order string, group string, limit int, start int) string {

	if b.table == "" {
		panic("没有指定表名")
	}

	queryString := fmt.Sprintf("SELECT %s FROM %s ", buildFields(fields), b.table)

	// 构建where条件
	cond := b.buildCond()
	if cond != "" {
		queryString += " WHERE " + cond
	}

	if group != "" {
		queryString += " GROUP BY " + group
	}

	if order != "" {
		queryString += " ORDER BY " + order
	}

	if limit > 0 && start > 0 {
		queryString += " LIMIT " + strconv.Itoa(start) + "," + strconv.Itoa(limit)
	} else if limit > 0 {
		queryString += " LIMIT " + strconv.Itoa(limit)
	}

	return queryString
}

// Multi 构造分页返回结果
// count int64 总数
// perpage int64 每页数量
// page int64 当前页
func Multi(count int64, page int, perpage int) ResultData {
	var data ResultData
	data.Count = int64(math.Max(float64(count), 0))                              // 总数
	data.Page = int(math.Max(float64(page), 1))                                  // 当前页
	data.PerPage = int(math.Max(float64(perpage), 1))                            // 每页数量
	data.PageCount = int(math.Ceil(float64(data.Count) / float64(data.PerPage))) // 总页数
	data.Start = int(math.Max(float64(data.PerPage*data.Page-data.PerPage), 0))  // 当前页之前有多少条数据
	data.Mark = data.Start + 1                                                   // 当前页开始是第几条数据
	return data
}

// Count 查询统计个数
func (b *Builder) Count() int64 {
	if b.table == "" {
		panic("没有指定表名")
	}
	data, err := b.FetchRow([]interface{}{Raw("COUNT(*) AS num")}, "", "", 0)
	if err != nil {
		return 0
	}

	return toInt64(data["num"])
}

// FetchWithPage 查询并返回多条记录，且包含分页信息
// field string 返回的字段 示例："*"
// order string 排序 示例："display_order DESC,id DESC"
// group string 分组字段 示例："user_group"
// limit 结果数量
// page 第几页，从1开始
func (b *Builder) FetchWithPage(fields []interface{}, order string, group string, limit int, page int) (*ResultData, error) {

	cond := b.buildCond()
	count := b.Clear().WhereRaw(cond).Count()
	data := Multi(count, page, limit)

	var err error = nil
	data.List, err = b.Clear().WhereRaw(cond).Fetch(fields, order, group, limit, data.Start)
	return &data, err
}

// Fetch 查询并返回多条记录
// field string 返回的字段 示例："*"
// order string 排序 示例："display_order DESC,id DESC"
// group string 分组字段 示例："user_group"
// limit 结果数量
// start 起始位置
func (b *Builder) Fetch(fields []interface{}, order string, group string, limit int, start int) ([]map[string]interface{}, error) {

	queryString := b.MakeQueryString(fields, order, group, limit, start)

	return b.schema.GetAll(queryString)

}

// FetchRow 查询并返回单条记录
// field string 返回的字段 示例："*"
// order string 排序 示例："display_order DESC,id DESC"
// group string 分组字段 示例："user_group"
// start 起始位置
func (b *Builder) FetchRow(fields []interface{}, order string, group string, start int) (map[string]interface{}, error) {

	queryString := b.MakeQueryString(fields, order, group, 1, start)

	return b.schema.GetRow(queryString)
}

// FetchOne 查询并返回单个字段
// field string 返回的字段 示例："count(*) AS count"
// order string 排序 示例："display_order DESC,id DESC"
// group string 分组字段 示例："user_group"
// start 起始位置
func (b *Builder) FetchOne(field interface{}, order string, group string, start int) string {

	queryString := b.MakeQueryString([]interface{}{field}, order, group, 1, start)

	item, err := b.schema.GetRow(queryString)
	if err != nil {
		panic(err)
	}

	return toString(item[toString(field)])
}

// Insert 单条插入
// set map[string]interface{} 插入的数据
// 返回插入的id，错误信息
func (b *Builder) Insert(set map[string]interface{}) (int64, error) {
	if b.table == "" {
		panic("没有指定表名")
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	// 拼接查询语句
	var fields []string
	var values []string
	for name, value := range set {
		fields = append(fields, buildK(name))
		values = append(values, "'"+Addslashes(toString(value))+"'")
	}
	queryString := "INSERT INTO " + b.table + " (" + strings.Join(fields, ",") + ") VALUES (" + strings.Join(values, ",") + ")"

	// 执行插入语句
	res, err := b.schema.Exec(queryString)
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
func (b *Builder) Update(set map[string]interface{}, limit ...int) (int64, error) {
	if len(limit) > 1 {
		panic("too many arguments")
	}

	if b.table == "" {
		panic("没有指定表名")
	}

	// 构建where条件
	where := b.buildCond()
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	if len(set) == 0 {
		return 0, ErrEmptyValue
	}

	// 查询字符串
	queryString := "UPDATE " + b.table + " SET " + buildVal(set, []string{}) + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	// 执行更新语句
	res, err := b.schema.Exec(queryString)
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
func (b *Builder) Increment(column string, amount int64, set ...map[string]interface{}) (int64, error) {
	if len(set) > 1 {
		panic("too many arguments")
	}
	if b.table == "" {
		panic("没有指定表名")
	}

	// 构建where条件
	where := b.buildCond()
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	column = buildK(column)
	// 拼接自增语句
	var extra []string
	if amount >= 0 {
		extra = []string{
			fmt.Sprintf("%s=%s+%d", column, column, amount),
		}
	} else {
		extra = []string{
			fmt.Sprintf("%s=%s-%d", column, column, -amount),
		}
	}

	// 拼接sql
	var queryString = ""
	if len(set) > 0 && set[0] != nil {
		queryString = "UPDATE " + b.table + " SET " + buildVal(mergeMap(set...), extra) + " WHERE " + where
	} else {
		queryString = "UPDATE " + b.table + " SET " + extra[0] + " WHERE " + where
	}

	// 执行更新sql语句
	res, err := b.schema.Exec(queryString)
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
func (b *Builder) Decrement(column string, amount int64, set ...map[string]interface{}) (int64, error) {
	return b.Increment(column, -amount, set...)
}

// Delete 删除
func (b *Builder) Delete(limit ...int) (int64, error) {
	if len(limit) > 1 {
		panic("too many arguments")
	}

	if b.table == "" {
		panic("没有指定表名")
	}

	// 构建where条件
	where := b.buildCond()
	if where == "" {
		// 防止在update、delete操作时，漏掉条件造成的严重后果
		// 如果确实不需要条件，请将条件设置为 1=1
		return 0, ErrEmptyCond
	}

	// 拼接删除语句
	queryString := "DELETE FROM " + b.table + " WHERE " + where

	if len(limit) == 1 && limit[0] > 0 {
		queryString += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	// 获取影响的行数
	res, err := b.schema.Exec(queryString)
	if err != nil {
		return 0, err
	}

	// 获取影响的行数
	return res.RowsAffected()
}
