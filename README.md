# lazygo

lazygo框架


## 快速开始


### 初始化lazygo项目

方案一：
```bash

# 安装lazygo
go install github.com/lazygo/lazygo@latest

# 创建项目 lazygo create 包名 目录名
lazygo create github.com/lazygo/myproject myproject
cd myproject

# 运行
make run

# 构建
make build

```

方案二：

```bash

# 安装lazygo
go install github.com/lazygo/lazygo@latest

# 创建项目目录
mkdir myproject
cd myproject

# 初始化项目 lazygo init 包名
lazygo init github.com/lazygo/myproject

# 运行
make run

# 构建
make build

```


### 项目文件目录结构说明

项目初始化后，将会自动创建下列文件目录结构。

```text
.
├── app                              // http请求处理，业务逻辑代码目录
│   ├── controller                   // 控制器
│   │   └── ...
│   ├── middleware                   // 中间件
│   │   └── ...
│   └── request                      // 定义请求和响应
│   │   └── ...
├── config                           // 配置初始化和组件配置注册，可根据需要增减此处代码
│   └── config.go
├── config.toml                      // 配置文件，在config中加载
├── framework                        // 框架相关，通常无需修改
│   ├── app.go                       // 提供Server单例
│   ├── context.go                   // 用户拓展server.Context，在controller和middleware中可以使用拓展Context中提供的方法
│   ├── exception.go                 // 错误处理
│   ├── handler.go                   // 提供一些 HandlerFunc 转换函数
│   └── logger.go                    // 提供日志记录器
├── go.mod
├── main.go
├── Dockerfile
├── Makefile
├── README.md
├── db.sql                           // 项目数据库
├── model                            // model层
│   ├── cache                        // 缓存model
│   │   └── ...
│   ├── db                           // sql数据库model
│   │   └── ...
│   └── model.go
├── pkg                              // 组件包，例如支付sdk组件、云服务sdk组件等
│   └── ...
├── router                           // 路由，需要注册路由后，才能通过http访问服务
│   ├── base.go                      // 基础路由，注册一些全局有效的中间件
│   └── ...                          // 业务路由，可以注册uri到控制器
└── utils                            // 工具
    ├── cache_key.go                 // 缓存key
    ├── define.go                    // 常量
    ├── errors                       // 错误定义
    │   └── errors.go
    └── string.go                    // 字符串处理函数

```

## 配置

项目的配置加载相关代码在`config/config.go`中，可根据实际使用情况修改此文件内容。

lazygo 默认支持toml和json两种格式的配置。

```go
// 加载toml格式的配置
loader := config.Toml(data []byte) (*Config, error)
// 加载json格式的配置
loader := config.Json(data []byte) (*Config, error)
```

可以调用func (c *Config) Load(field string f any) 方法，将配置信息中的field配置段解析到回调函数f的第一个参数中。
其中回调函数f的定义需符合 func(*CustomStruct) error 或 func([]CustomStruct) error 的形式。 注意需要确保提供的配置内容与回调函数参数结构体类型CustomStruct的数据字段相匹配。如果加载的是json类型的配置，则需要在CustomStruct结构体字段中提供json注解，同理toml格式配置需要提供toml注解。
以下代码为解析配置redis配置段到RedisConfig结构体的示例。

示例：加载json配置
```go
import (
    "github.com/lazygo/lazygo/config"
)

// RedisConfig 可同时增加json和toml注解，便于更换配置文件格式。
type RedisConfig struct {
	Name     string `json:"name" toml:"name"`
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Password string `json:"password" toml:"password"`
	Db       int    `json:"db" toml:"db"`
	Prefix   string `json:"prefix" toml:"prefix"`
}


// 加载json配置
func LoadJsonConfig() {
    jsonConfig := `
    {
        "redis": [
            {
                "name": "redis_1",
                "host": "127.0.0.1",
                "port": 6379,
                "password": "",
                "db": 0
            },
            {
                "name": "redis_2",
                "host": "127.0.0.1",
                "port": 6380,
                "password": "",
                "db": 0
            }
        ],
	"redis": {
             "foo": "bar"
	}
    }
    `

    // 使用json加载器解析配置内容
    jsonLoader, err := config.Json(data)
	if err != nil {
		return err
	}

	// 注册redis配置
	err = jsonLoader.Load("redis", func(conf []RedisConfig) error {
		fmt.Pringln(conf)
	})
	if err != nil {
		return err
	}

}

```

## 路由

## 控制器

### 控制器的定义 ###

控制器结构体必须包含类型为Context的Ctx成员变量。其定义如下所示：

```go
// UserController 定义一个User控制器
type UserController struct{
	Ctx Context
}
```

### 控制器方法的定义规则 ###

控制器的Public类型成员函数用于处理业务逻辑，其入参和出参定义形式必须遵循以下规则。

- 入参必须有且仅有一个，且此参数必须实现server.Request接口。
- 返回参数可以有一个或两个，且最后一个返回参数必须时error类型。
- 当返回仅有1个error类型参数时，如果返回的error为nil框架将不会向HTTP输出任何内容。这可以方便开发者在控制器方法中自行定义HTTP响应的内容。如果返回error不为nil，则根据`framework/exception.go`中定义的格式输出错误信息到HTTP响应。
- 当返回有2个参数时，如果返回的第二个参数error为nil，则会根据server.HTTPOKHandler中定义的格式，输出第一个参数到HTTP响应。如果返回error不为nil，处理方式与1个参数时相同，直接向HTTP响应输出错误信息。

```go
// 自定义响应内容的控制器方法
func (c *UserController)Login(request.UserLoginRequest) error {
    ctl.Ctx.HTMLBlob(200, []byte("<h1>lazygo framework</h1>"))
    return nil
}

// 返回一个request.UserProfileResponse结构，框架默认会使用ctx.Succ() 处理返回的*request.UserProfileResponse
// 也就是会将第一个参数*request.UserProfileResponse 整个放入"data"字段中，并序列化成json。{"data": {"name": "李某人"}}
func (c *UserController)Profile(request.UserProfileRequest) (*request.UserProfileResponse, error) {
    resp := &request.UserProfileResponse{
        Name: "李某人",
    }
    return resp, nil
}

// 返回一个map结构，返回结果处理方式同上{"data": {"lazygo": "yes"}}
func (c *UserController)Profile(request.UserProfileRequest) (interface{}, error) {
    resp := map[string]interface{}{
        "lazygo": "yes",
    }
    return resp, nil
}

// 返回一个error信息，返回结果类似 {"errno": 400, "msg": "参数错误"}
// 可在`framework/exception.go`中自行定义返回结果的渲染格式
func (c *UserController)Profile(request.UserProfileRequest) (interface{}, error) {
    return nil, server.NewHTTPError(200, 400, "参数错误")
}
```

需要注意的是，框架会默认认为所有Public成员函数都是HTTP请求处理函数，也就是说这些Public类型成员函数都必须遵循上述入参和出参的定义规则。在框架启动时会强制检查该控制器下所有的Public成员函数是否满足此规则。如果有不符合此规则的函数，请定义为开头小写的私有类型。
例如`(ctl *UserController) Register(req *request.User) (any, error)`和`(ctl *UserController) Login(req *request.User) (any, error)`两个函数内部都会调用发送通知的方法，可将发送通知定义为`(ctl *UserController)sendMsg(uid uint64, msg string)`并在Register和Login中以`ctl.sendMsg(uid, msg)`的形式调用。

### 注册控制器到HTTP路由 ###

控制器函数需要注册到路由才能被HTTP请求访问到，注册路由时需要使用`server.Controller`函数 将控制器函数转换为路由中的HandlerFunc。


```go
// server.Controller 函数定义如下
// 参数h为控制器结构体实例
// method为函数名，若不指定函数名，会自动指定函数名为路由uri最后一个“/”后面的字符串
func Controller(h interface{}, methodName ...string) server.HandlerFunc
```

注册路由示例代码：

```go
// 将uri /user/profile 注册到UserController.UserInfo
app.Post("/user/profile", server.Controller(controller.UserController{}, "UserInfo"))

// 将uri /user/login 注册到UserController.Login
app.Post("/user/login", server.Controller(controller.UserController{}))
```

### Request 参数绑定和预处理 ###

```go
type ToolsUploadRequest struct {
    Category string      `json:"category" bind:"query,form" process:"trim,cut(20)"`
    Tags     []int       `json:"tag" bind:"query,form"`
    Image    server.File `json:"image"``
}
```

在框架收到HTTP请求时，请求数据会自动绑定到Request结构体中。绑定需要依赖结构体注解来完成。

- 对于 `Content-Type`为 `application/json` 的请求，会自动将json数据解析到结构体中。

- 对于 `Content-Type`为 `form-data`类型的请求或GET请求，可通过 `bind` 注解 指定绑定的数据来源。例如 bind:"query,form" 表示优先从url的query参数中获取字段，如果获取不到，则使用form获取。

- bind支持的类型为：
  
    `value` 或 `ctx` 表示调用context中Value方法获取数据；
    `header` 表示从HTTP Header中获取数据；
    `param`从路由参数中获取数据；
    `query` 从URL Query中获取数据；
    `form` 从Post Form中获取数据；
    `file` 绑定文件数据。

- 数据类型为切片时，需要提供的参数格式为 使用逗号分隔的字符串，或json数组字符串。例如`?tags=1,2,3` 或`?tags=[1,2,3]`都可以绑定到`[]int`类型
    
在绑定数据时，会自动通过注解中的预处理器函数，对被绑定的参数进行预处理。多个预处理器之间使用逗号隔开。

`trim` ：剔除字符串两端空字符；
`tolower` ：字符串转为小写；
`toupper` ：字符串转为大写；
`cut(num int)` ：如果utf8字符串字符数量超过num个，则将字符串截断。


`Request`需要实现 `func Verify() error` 和 `func Clear()` 两个函数

- `Verify`函数在参数绑定后执行，用于对请求参数内容进行校验

- `Clear`会在http请求返回响应后执行，用于做一些清理工作


```go
func (r *ToolsUploadRequest) Verify() error {

	if slices.Contains(utils.ImageFormat, path.Ext(r.Image.FileHeader.Filename)) == false {
		return errors.ErrInvalidImageFormat
	}
	return nil
}

func (r *ToolsUploadRequest) Clear() {
	if r.Image.File != nil {
		r.Image.File.Close()
	}
}
```


## 中间件

## 上下文Context

## 日志

## 模型

### mysql

#### 连接数据库

数据库驱动需在使用前自行引入：

```go
import _ "github.com/go-sql-driver/mysql"
```

```go
err := sqldb.Init([]sqldb.MySQLConfig{
    {
        Driver:          "mysql",     // 数据库驱动
        Name:            "default",   // 连接名称，通过 sqldb.Database("default") 获取
        User:            "root",
        Passwd:          "",
        Host:            "127.0.0.1",
        Port:            3306,
        DbName:          "mydb",
        MaxOpenConns:    50,
        MaxIdleConns:    10,
        ConnMaxLifetime: 300,         // 连接最大存活时间，单位秒
        Params:          map[string]string{"charset": "utf8mb4", "parseTime": "true"},
    },
})
if err != nil {
    return err
}

db, err := sqldb.Database("default")

// 程序退出时关闭全部连接
defer sqldb.CloseAll()
```

注意：扫描 `time.Time` 类型的字段需要在 `Params` 中设置 `parseTime=true`。

#### 查询构建器

`db.Table()` 每次返回一个新的查询构建器。构建器不是 goroutine 安全的，链式调用时暂存的条件、ORDER BY、LIMIT 等状态仅属于本次串式操作。

**查询**

```go
// Where 自动识别条件类型，可多次链式调用，多个条件之间以 AND 连接
var users []User
_, err := db.Table("user").
    Where("name", "张三").                  // `name` = ?
    Where("age", ">", 18).                  // `age` > ?
    Where("status", "in", []any{1, 2}).     // `status` IN(?, ?)
    Where("deleted_at", nil).               // `deleted_at` IS NULL
    WhereMap(map[string]any{"age >=": 20}). // `age` >= ?
    OrderBy("id", "desc").
    Limit(10).
    Offset(20).
    Find(&users)
```

**Where 条件的所有形式**

`Where` 根据参数个数与类型自动识别条件：

```go
db.Table("user").
    // —— 1个参数 ——
    Where("age > 0").                       // 字符串：原生条件子句，原样拼入SQL（注意自行保证安全）
    Where(map[string]any{"name": "a"}).     // map：等价于 WhereMap

    // —— 2个参数 ——
    Where("name", "张三").                   // 普通值：等于，生成 `name` = ?
    Where("age >", 18).                     // key为"字段 运算符"：按运算符解析，同 WhereMap 的key规则
    Where("deleted_at", nil).               // 值为nil：判空，生成 `deleted_at` IS NULL
    Where("id", []int{1, 2, 3}).            // 值为任意类型切片：IN查询，生成 `id` IN(?, ?, ?)

    // —— 3个参数 ——
    Where("age", ">", 18).                  // 指定运算符：运算符原样拼入，如 > >= < <= != like 等
    Where("deleted_at", "=", nil).          // 值为nil：= 生成 IS NULL，!= 生成 IS NOT NULL
    Where("status", "in", []any{1, 2}).     // 运算符 in / not in（不区分大小写）：IN / NOT IN 查询，
                                            // 值为非切片时按单元素IN处理
```

除 `Where` 外还提供了专用条件方法：

```go
WhereIn("id", ids)                          // IN查询；传空切片时生成恒假条件 1=0
WhereNotIn("id", ids)                       // NOT IN查询；传空切片时生成恒真条件 1=1
WhereRaw("age > ? AND name != ?", 18, "x")  // 原生条件子句，? 与 args 按位置对应
ClearCond()                                 // 手动清空已暂存的条件
```

**WhereMap 规则**

- key 不含空格时按等于处理；值为 nil 生成 `IS NULL`，值为切片生成 IN 查询
- key 含空格时按空格拆分为 `字段 运算符`，运算符支持 `>` `>=` `<` `<=` `!=` `in` `not in` `like`
- 值为 map 时，以该 key（AND/OR）为关系符嵌套条件分组：

```go
// ((`name` = 'a' AND `age` > 1) OR `name` = 'b') AND `status` = 1
db.Table("user").Where(map[string]any{
    "OR": map[string]any{
        "AND":  map[string]any{"name": "a", "age >": 1},
        "name": "b",
    },
    "status": 1,
})
```

**排序、分组与分页**

```go
// 排序：方向仅支持 ASC / DESC（不区分大小写），可多次调用按顺序拼接
db.Table("user").Where("age", ">", 18).OrderBy("age", "desc").OrderBy("id", "asc").Find(&users)

// 随机排序
db.Table("user").OrderByRand().Find(&users)

// 指定查询字段：不调用Select时默认 SELECT *；Select 会按字段名排序后拼接
_, err := db.Table("user").Select("id", "name").Where("id", 1).First(&user)

// 分组：配合 Select 中的聚合函数做分组统计
type AgeStat struct {
    Age   int64 `json:"age"`
    Total int64 `json:"total"`
}
var stats []AgeStat
_, err = db.Table("user").
    Select("age", "count(*) AS total").
    GroupBy("age").
    Find(&stats)

// 分页：Limit 限制条数；Offset 为偏移量，需与 Limit 同时使用才生效
var page []User
_, err = db.Table("user").Where("status", 1).OrderBy("id", "desc").Limit(10).Offset(20).Find(&page)
```

注意：不要对分组查询使用 `Count()`——`Count` 是单值查询，与 `GroupBy` 同用时不统计组数。分组统计请使用 `Select` + 聚合函数 + `Find` 的方式（如上面的 AgeStat 示例）。

**结果绑定**

- struct / []struct：按 json tag 匹配列名；匿名字段平铺；有 tag 的嵌套结构体以 tag 为前缀（如 `U.id`）；tag 带选项（`json:"uid,omitempty"`）也能正确匹配；`sql.NullString` 等 database/sql 类型直接作为叶子字段匹配
- `map[string]string` / `map[string]any`：值为字符串类型

**单条、单字段与统计**

```go
var user User
n, err := db.Table("user").Where("id", 1).First(&user)
// n == 0 表示记录不存在，err 为 nil

// 查询单个字段，支持表达式与别名
name, err := db.Table("user").Where("id", 1).One("name")
count, err := db.Table("user").Where("age", ">", 18).One("count(*) AS count")

total, err := db.Table("user").Where("age", ">", 18).Count()

// 只拼接SQL不执行。QueryString 为纯查询，不会清空构建器状态
sql, args, err := db.Table("user").Where("id", 1).QueryString()
```

**写入**

```go
// 插入，返回自增id
id, err := db.Table("user").Insert(map[string]any{"name": "a", "age": 1})

// 更新，返回影响行数；可选第二个参数限制更新行数
n, err := db.Table("user").Where("id", 1).Update(map[string]any{"name": "b"}, 1)

// 原生SET更新
n, err = db.Table("user").Where("id", 1).UpdateRaw("age = age + 1, name = 'c'")

// 自增/自减，可同时更新其他字段
n, err = db.Table("user").Where("id", 1).Increment("age", 1)
n, err = db.Table("user").Where("id", 1).Decrement("age", 2, map[string]any{"name": "d"})

// 删除，返回影响行数
n, err = db.Table("user").Where("id", 1).Delete()
```

防止误操作：`Update` / `UpdateRaw` / `Increment` / `Delete` 必须携带 WHERE 条件，否则返回 `ErrEmptyCond`。
确实需要全表操作时，请显式传入 `Where("1=1")`。

**事务**

`Transaction` 中返回 error 或 panic 都会回滚；panic 会以 error 形式返回。

```go
err := db.Transaction(func(tx *sqldb.Tx) error {
    id, err := tx.Table("user").Insert(map[string]any{"name": "a"})
    if err != nil {
        return err
    }
    _, err = tx.Table("user").Where("id", id).Update(map[string]any{"age": 1})
    return err
})
```

**表关联**

```go
table := sqldb.Table("user").As("U").
    LeftJoin(sqldb.Table("order").As("O"), "U.id", "O.user_id")

var list []UserOrder
_, err := db.Table(table.String()).
    Where("U.id", 1).
    Select(sqldb.StructFields(UserOrder{})...).
    Find(&list)
```

**执行前 Hook**

`BeforeHook` 接收一个执行前回调，其返回值在SQL执行后调用（类似 defer）。hook 仅对构建器的下一次执行生效。

```go
builder := db.Table("user").BeforeHook(func(query string, args ...any) func() {
    start := time.Now()
    return func() {
        fmt.Println("SQL耗时:", time.Since(start), query, args)
    }
})
builder.Where("id", 1).Find(&users)
```

**构建器状态说明**

- 每个终端方法（`Find` / `First` / `One` / `Count` / `Insert` / `Update` / `UpdateRaw` / `Increment` / `Delete`）执行后会自动清空条件、ORDER BY、LIMIT 等暂存状态，builder 可继续复用
- `QueryString` 是纯拼接，不清理任何状态

**原生SQL**

```go
var users []User
n, err := db.Find(&users, "SELECT * FROM user WHERE age > ?", 18)

var user User
n, err = db.First(&user, "SELECT * FROM user WHERE id = ?", 1)

result, err := db.Exec("UPDATE user SET name = ? WHERE id = ?", "a", 1)
```

#### Model 封装

`TxModel[T]` 以泛型结构体约定封装常用 CRUD，数据约定：

- 主键 `id` 使用 `uint64`，时间戳字段使用 `int64`
- 创建/更新时间默认不自动写入，通过 `AutoCTime`/`AutoMTime` 指定字段名后自动维护：`Create` 时写入创建时间和更新时间，`Modify` 时写入更新时间
- 数据库字段定义为 NOT NULL；允许 NULL 的字段使用 `sql.Null[T]` 类型
- 结构体字段通过 json tag 与列名匹配

```go
type UserData struct {
    Id    uint64         `json:"id"`
    Name  string         `json:"name"`
    Email sql.NullString `json:"email"`
    Ctime int64          `json:"ctime"`
    Mtime int64          `json:"mtime"`
}

type UserModel struct {
    sqldb.TxModel[UserData]
}

func NewUserModel() *UserModel {
    mdl := &UserModel{}
    mdl.SetTable("user")
    db, _ := sqldb.Database("default")
    mdl.SetTx(&db.Tx)
    return mdl
}

// 开启自动维护创建/更新时间（不调用则默认不写入，字段名可自定义）
mdl.AutoCTime("ctime")
mdl.AutoMTime("mtime")
```

使用示例：

```go
mdl := NewUserModel()

// 新增，自动写入ctime/mtime，返回自增id
id, err := mdl.Create(&UserData{Name: "a"})

// 按主键更新（默认id），mtime自动刷新
n, err := mdl.Modify(&UserData{Id: uint64(id), Name: "b"})
// 也可指定其他唯一键作为更新条件
n, err = mdl.Modify(&UserData{Uid: 100, Name: "b"}, "uid")

// 查询：空结果返回 nil, 0, nil，需同时判断 err 和 n
data, n, err := mdl.First(map[string]any{"id": id})
if err != nil {
    return err
}
if n == 0 {
    // 记录不存在
}

list, n, err := mdl.Find(map[string]any{"age >": 18})

ok, err := mdl.Exists(map[string]any{"name": "a"})
cnt, err := mdl.Count(map[string]any{"age >": 18})

// 分页：Order/Sort为主排序字段，Orders可附加次级排序，groupby传nil表示不分组
req := &sqldb.PaginatorRequest{
    Page:  1,    // 页码，从1开始
    Size:  10,   // 每页条数
    Order: "id", // 排序字段
    Sort:  "desc",
    Orders: []sqldb.OrderBy{ // 可选：次级排序
        {Field: "age", Sort: "asc"},
    },
}
resp, n, err := mdl.Paginator(map[string]any{"status": 1}, req, nil)
// resp.Page / resp.Size / resp.Total / resp.List
// Total 为总记录数；入参可先经 Clean 清洗（非法page/size回落默认值，size超上限截断）
req.Clean(sqldb.DefaultPaginatorMaxSize, sqldb.DefaultPaginatorRequest)

// 分组分页：total统计组数；不传fields时默认只查询分组字段
resp, _, err = mdl.Paginator(map[string]any{}, req, []string{"age"})
```

在事务中使用：在 `db.Transaction` 的回调里重新 `SetTx(tx)` 即可让 Model 的操作加入事务。

```go
err := db.Transaction(func(tx *sqldb.Tx) error {
    mdl := NewUserModel()
    mdl.SetTx(tx)
    _, err := mdl.Create(&UserData{Name: "a"})
    return err
})
```

### 缓存

### 错误处理
