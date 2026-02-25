package model

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/examples/framework"
	"github.com/lazygo/lazygo/httpclient"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
	"github.com/lazygo/pkg/goutils"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrRequestApiFail = errors.New("request api fail")
	ErrRequestFail    = errors.New("request fail")
	ErrInvaildContent = errors.New("response content error")
)

var (
	DefaultPaginatorRequest = PaginatorRequest{Page: 1, Size: 10, Order: "id", Sort: "asc"}
	DefaultPaginatorMaxSize = 1000
)

type BaseResp struct {
	Code  int    `json:"code"`
	Errno int    `json:"errno"`
	Msg   string `json:"msg"`
	Rid   uint64 `json:"rid"`
	Time  int64  `json:"t"`
}

type Order struct {
	Field string `json:"field" bind:"query,form"`
	Sort  string `json:"sort" bind:"query,form"`
}

type PaginatorRequest struct {
	Order  string  `json:"order" bind:"query,form"`
	Sort   string  `json:"sort" bind:"query,form"`
	Orders []Order `json:"orders" bind:"query,form"`
	Page   int     `json:"page" bind:"query,form"`
	Size   int     `json:"size" bind:"query,form"`
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

func DB(dbname string) *mysql.DB {
	database, err := mysql.Database(dbname)
	if err != nil {
		panic(err)
	}
	return database
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
	tx    *mysql.Tx
}

func (m *TxModel[T]) SetTable(table string) {
	m.table = table
}

func (m *TxModel[T]) SetTx(tx *mysql.Tx) {
	m.tx = tx
}

func (m *TxModel[T]) SetDB(dbname string) {
	m.tx = &DB(dbname).Tx
}

func (m *TxModel[T]) Tx() *mysql.Tx {
	return m.tx
}

func (m *TxModel[T]) QueryBuilder() mysql.Builder {
	table := m.table
	if table == "" {
		// 没有指定表名
		panic("没有指定表名")
	}
	return m.tx.Table(table)
}

func (m *TxModel[T]) Table(table string) mysql.Builder {
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
		req.Orders = append([]Order{
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
	now := time.Now().Unix()
	set["ctime"] = now
	set["mtime"] = now
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

type CacheModel struct {
	cache.Cache
}

func (m *CacheModel) SetCache(name string) {
	m.Cache = m.cache(name)
}

func (m *CacheModel) cache(name string) cache.Cache {
	instance, err := cache.Instance(name)
	if err != nil {
		panic(fmt.Sprintln(name, err))
	}
	return instance
}

type RedisModel struct {
	*goredis.Client
}

func (m *RedisModel) SetClient(name string) {
	m.Client = m.client(name)
}

func (m *RedisModel) client(name string) *goredis.Client {
	instance, err := redis.Client(name)
	if err != nil {
		panic(fmt.Sprintln(name, err))
	}
	return instance
}

type RPCModel struct {
	BaseURL string
	IPs     []netip.Addr
}

var (
	client     *httpclient.Client
	clientOnce sync.Once
)

func DefaultClient() *httpclient.Client {
	clientOnce.Do(func() {
		client = httpclient.New(&httpclient.Config{
			// DNSResolverAddr:"",
			HTTPDNSAdapter: "baidu",
		}).Client(&httpclient.HttpConfig{
			Timeout:             10 * time.Second,
			MaxIdleConnsPerHost: 20,
		})
	})
	return client
}

func (mdl *RPCModel) Request(ctx framework.Context, method string, uri string, data []byte, v any) error {
	url := strings.TrimRight(mdl.BaseURL, "/") + "/" + strings.TrimLeft(uri, "/")

	headers := map[string]string{"Authorization": "sign"}

	var ips []string
	for _, ip := range mdl.IPs {
		ips = append(ips, ip.String())
	}
	if len(ips) > 0 {
		headers[httpclient.HeaderSpecifiedIP] = strings.Join(ips, ",")
	}

	ctx.Logger().Debug("rpc request: %s %s %v", method, url, headers)

	body, status, err := DefaultClient().Request(ctx, method, url, data, headers)
	ctx.Logger().Debug("rpc response: %s", string(body))
	if err != nil {
		return errors.Join(ErrRequestFail, err)
	}
	var resp struct {
		BaseResp
		Data json.RawMessage `json:"data"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return errors.Join(ErrInvaildContent, err)
	}
	if resp.Errno != 0 {
		return fmt.Errorf("%w: %s", ErrRequestApiFail, resp.Msg)
	}
	if status != http.StatusOK {
		return fmt.Errorf("%w: %s", ErrRequestApiFail, "network error")
	}
	if v == nil {
		return nil
	}
	return json.Unmarshal(resp.Data, v)
}
