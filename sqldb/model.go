package sqldb

import (
	"cmp"
	"strings"
	"time"

	"github.com/lazygo/pkg/goutils"
)

var (
	DefaultPaginatorRequest = PaginatorRequest{Page: 1, Size: 10, Order: "id", Sort: "asc"}
	DefaultPaginatorMaxSize = 1000
)

type OrderBy struct {
	Field string `json:"field" bind:"query,form"`
	Sort  string `json:"sort" bind:"query,form"`
}

type PaginatorRequest struct {
	Order  string    `json:"order" bind:"query,form"`
	Sort   string    `json:"sort" bind:"query,form"`
	Orders []OrderBy `json:"orders" bind:"query,form"`
	Page   int       `json:"page" bind:"query,form"`
	Size   int       `json:"size" bind:"query,form"`
}

func (r *PaginatorRequest) Clean(maxSize int, def PaginatorRequest) {
	if r.Page < 1 {
		r.Page = cmp.Or(def.Page, 1)
	}
	if r.Size < 1 {
		r.Size = cmp.Or(def.Size, 10)
	}
	if r.Size > maxSize {
		r.Size = maxSize
	}
	if r.Order == "" {
		r.Order = cmp.Or(def.Order, "id")
	}
	if r.Sort == "" {
		r.Sort = cmp.Or(def.Sort, "asc")
	}
}

type PaginatorResponse[T any] struct {
	Page  int   `json:"page"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

// TxModel 数据库model
// T ModelData 结构体定义规范:
//   - id使用 uint64
//   - 时间使用 int64
//   - 创建时间字段位ctime，更新时间字段位mtime
//   - 数据库定义字段必须为NOT NULL，对于特殊情况下允许为NULL的字段，ModelData结构体对应的字段需使用sql.Null[T]字段类型
//
// 具体XxxModel 定义规范：
//   - XxxModel 结构体必须嵌入DBModel[ModelData]
//     type UserModel struct {
//     model.TxModel[UserData]
//     }
//   - XxxModel 的接收器为(mdl *XxxModel)
//   - Controller中使用XxxModel时，变量名统一为为mdlXxx
type TxModel[T any] struct {
	table string
	tx    *Tx
}

func (m *TxModel[T]) SetTable(table string) {
	m.table = table
}

func (m *TxModel[T]) SetTx(tx *Tx) {
	m.tx = tx
}

func (m *TxModel[T]) SetDB(dbname string) error {
	db, err := Database(dbname)
	if err != nil {
		return err
	}
	m.tx = &db.Tx
	return nil
}

func (m *TxModel[T]) Tx() *Tx {
	return m.tx
}

func (m *TxModel[T]) QueryBuilder() Builder {
	table := m.table
	if table == "" {
		// 没有指定表名
		panic("没有指定表名")
	}
	return m.tx.Table(table)
}

func (m *TxModel[T]) Table(table string) Builder {
	if table == "" {
		// 没有指定表名
		panic("没有指定表名")
	}
	return m.tx.Table(table)
}

func (mdl *TxModel[T]) First(cond map[string]any, fields ...string) (*T, int, error) {
	var data T
	n, err := mdl.QueryBuilder().Where(cond).Select(fields...).First(&data)
	if err != nil {
		return nil, 0, err
	}
	return &data, n, nil
}

func (mdl *TxModel[T]) Exists(cond map[string]any) (bool, error) {
	data := map[string]any{}
	n, err := mdl.QueryBuilder().Where(cond).Select("(0)").First(&data)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (mdl *TxModel[T]) Count(cond map[string]any) (int64, error) {
	return mdl.QueryBuilder().Where(cond).Count()
}

func (mdl *TxModel[T]) Find(cond map[string]any, fields ...string) ([]T, int, error) {
	var data []T
	n, err := mdl.QueryBuilder().Where(cond).Select(fields...).Find(&data)
	if err != nil {
		return nil, 0, err
	}
	return data, n, nil
}

func (mdl *TxModel[T]) Paginator(cond map[string]any, req *PaginatorRequest, groupby []string, fields ...string) (*PaginatorResponse[T], int, error) {
	var list []T
	builder := mdl.QueryBuilder().Where(cond)
	if req.Order != "" {
		sort := "asc"
		if strings.ToLower(req.Sort) == "desc" {
			sort = "desc"
		}
		req.Orders = append([]OrderBy{
			{
				Field: req.Order,
				Sort:  sort,
			},
		}, req.Orders...)
	}
	for _, order := range req.Orders {
		builder.OrderBy(order.Field, order.Sort)
	}
	if len(groupby) > 0 {
		builder.GroupBy(groupby...)
	}
	n, err := builder.
		Offset(int64(req.Size * max((req.Page-1), 0))).
		Limit(int64(req.Size)).
		Select(fields...).Find(&list)
	if err != nil {
		return nil, 0, err
	}
	count, err := mdl.QueryBuilder().Where(cond).Count()
	if err != nil {
		return nil, 0, err
	}
	data := &PaginatorResponse[T]{
		Page:  req.Page,
		Size:  req.Size,
		Total: count,
		List:  list,
	}
	return data, n, nil
}

func (mdl *TxModel[T]) Insert(set map[string]any) (int64, error) {
	now := time.Now().Unix()
	set["ctime"] = now
	set["mtime"] = now
	return mdl.QueryBuilder().Insert(set)
}

func (mdl *TxModel[T]) Create(data *T, pk ...string) (int64, error) {
	if len(pk) == 0 {
		pk = []string{"id"}
	}
	set, _ := goutils.Struct2Map(data, pk...)
	return mdl.Insert(set)
}

func (mdl *TxModel[T]) Modify(data *T, pk ...string) (int64, error) {
	if len(pk) == 0 {
		pk = []string{"id"}
	}
	set, pkset := goutils.Struct2Map(data, pk...)
	now := time.Now().Unix()
	set["mtime"] = now
	return mdl.QueryBuilder().Where("id", pkset["id"]).Update(set)
}
