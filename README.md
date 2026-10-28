# lazygo

lazygo 是一个 Go 语言 Web 框架，提供路由、控制器、模型层（MySQL/缓存/分布式锁）等常用组件，帮助快速搭建 HTTP 服务。

## 目录

- [快速开始](#快速开始)
  - [初始化lazygo项目](#初始化lazygo项目)
  - [项目文件目录结构](#项目文件目录结构)
- [配置](#配置)
- [路由](#路由)
  - [服务实例与启动](#服务实例与启动)
  - [注册路由](#注册路由)
  - [路径参数](#路径参数)
  - [路由组](#路由组)
  - [中间件](#中间件)
  - [静态文件](#静态文件)
  - [其他说明](#其他说明)
- [控制器](#控制器)
  - [控制器的定义](#控制器的定义)
  - [控制器方法的定义规则](#控制器方法的定义规则)
  - [注册控制器到HTTP路由](#注册控制器到http路由)
  - [Request 参数绑定和预处理](#request-参数绑定和预处理)
- [上下文Context](#上下文context)
- [日志](#日志)
- [模型](#模型)
  - [mysql](#mysql)
  - [缓存](#缓存)
  - [分布式锁](#分布式锁)
  - [错误处理](#错误处理)
- [动态库（c-shared）示例](#动态库c-shared示例)

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

### 项目文件目录结构

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

项目的配置加载相关代码在 `config/config.go` 中，可根据实际使用情况修改此文件内容。

lazygo 默认支持 toml 和 json 两种格式的配置。

```go
// 加载toml格式的配置
loader, err := config.Toml(data []byte) (*Config, error)
// 加载json格式的配置
loader, err := config.Json(data []byte) (*Config, error)
```

可以调用 `func (c *Config) Load(field string, f any)` 方法，将配置信息中的 field 配置段解析到回调函数 f 的第一个参数中。

其中回调函数 f 的定义需符合 `func(*CustomStruct) error` 或 `func([]CustomStruct) error` 的形式。注意需要确保提供的配置内容与回调函数参数结构体类型 CustomStruct 的数据字段相匹配。如果加载的是 json 类型的配置，则需要在 CustomStruct 结构体字段中提供 json 注解，同理 toml 格式配置需要提供 toml 注解。

以下代码为解析配置中 redis 配置段到 RedisConfig 结构体的示例。

示例：加载json配置

```go
import (
    "fmt"

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
func LoadJsonConfig(data []byte) error {
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
        ]
    }
    `
    _ = jsonConfig

    // 使用json加载器解析配置内容
    jsonLoader, err := config.Json(data)
    if err != nil {
        return err
    }

    // 注册redis配置
    err = jsonLoader.Load("redis", func(conf []RedisConfig) error {
        fmt.Println(conf)
        return nil
    })
    if err != nil {
        return err
    }
    return nil
}
```

## 路由

### 服务实例与启动

通过 `server.New()` 创建服务实例；业务项目通常在 `framework` 包中以单例方式维护（可参考 examples 项目的 `framework/app.go`）。

```go
app := server.New()

// 启动（阻塞）
err := app.Start(ctx, ":8080")

// 优雅停止
err := app.Shutdown(ctx)

// 立即关闭
err := app.Close()
```

### 注册路由

所有注册方法在 `*Server` 和 `*Group` 上均可用，统一签名为 `func(path string, h HandlerFunc, m ...MiddlewareFunc)`：path 为路由路径，h 为处理函数，最后的可变参数是**路由级中间件**（仅对当前路由生效）。

**按 HTTP 方法注册**

| 注册方法 | HTTP 方法 | 典型用途 |
|---|---|---|
| `Get` | GET | 查询资源 |
| `Post` | POST | 创建资源、表单提交 |
| `Put` | PUT | 整体更新资源 |
| `Patch` | PATCH | 局部更新资源 |
| `Delete` | DELETE | 删除资源 |
| `Head` | HEAD | 仅获取响应头 |
| `Options` | OPTIONS | 预检请求、查询支持的方法 |
| `Connect` | CONNECT | 建立隧道（如代理） |
| `Trace` | TRACE | 链路回显诊断 |

**特殊方法**

以下两个不是标准 HTTP 方法，而是**替代传输层**：让 WebSocket 连接和进程内调用复用框架的路由与控制器体系。客户端发出的每条消息会被转换为 method=WEBSOCKET（或 CALL）的内部请求，按消息中的 `uri` 走路由分发，响应再按 `rid` 配对送回。

| 注册方法 | 方法名 | 说明 |
|---|---|---|
| `WebSocket` | `WEBSOCKET` | 处理经 WebSocket 连接传入的请求（长连接 RPC） |
| `Call` | `CALL` | 处理进程内调用桥传入的请求（同进程 RPC） |

框架还内置 `PROPFIND`、`REPORT` 两个 WebDAV 方法常量（`server.MethodPropfind` / `server.MethodReport`），没有对应的便捷方法，可用 `Add` 注册。

**通用注册入口**

```go
// 一次性注册所有支持的方法（含上面的全部方法以及 PROPFIND/REPORT/WEBSOCKET/CALL）
app.Any(path, handler, middleware...)

// 注册多个指定方法
app.Match([]string{http.MethodGet, http.MethodPost}, path, handler, middleware...)

// 按方法名注册的底层入口
app.Add(method, path, handler, middleware...)
```

注意：`Add` 的 `method` 为大小写敏感的字符串，须使用标准大写形式（如 `http.MethodGet` 即 `"GET"`），小写或非标准名将无法匹配到处理器。

**注册示例**

```go
// 控制器方法（推荐，见[控制器](#控制器)一章）
app.Get("/user/profile", server.Controller(controller.UserController{}, "Profile"))

// 直接编写 HandlerFunc
app.Get("/ping", func(c server.Context) error {
    return c.Blob(200, server.MIMETextPlainCharsetUTF8, []byte("pong"))
})

// 带路由级中间件
app.Get("/connection/:token",
    server.Controller(controller.CommonController{}, "Connection"),
    middleware.User, middleware.AuthUser)
```

同一路径可以为不同方法分别注册处理器（按方法分槽存储，互不冲突）；同一路径同一方法重复注册时，后注册的会覆盖先前的。

### 路径参数

路由路径支持 `:name` 参数和 `*` 通配符，匹配优先级为 **静态 > 参数 > 通配**：

```go
app.Get("/user/:id", handler)      // 匹配 /user/123，参数通过 c.Param("id") 获取
app.Get("/assets/*", handler)      // 匹配 /assets/ 下的任意路径，c.Param("*") 获取剩余部分
```

### 路由组

`Group` 将一组路由挂在统一前缀下，并可附加组级中间件；组支持嵌套，子组自动继承父组的前缀和中间件：

```go
func InitRouter(app *server.Server) {
    // 全局中间件（见下一节）
    app.Use(middleware.RequestID)

    api := app.Group("/api", middleware.Auth)        // 组级中间件对组内全部路由生效
    rest := api.Group("/rest")                        // 最终前缀 /api/rest
    rest.Get("/profile", server.Controller(controller.UserController{}))

    audit := rest.Group("audit")                      // 嵌套组，前缀 /api/rest/audit
    audit.Get("/list", server.Controller(controller.AuditController{}))
}

// 也可在函数中按组拆分注册逻辑
func RestRouter(g *server.Group) {
    g = g.Group("", middleware.User, middleware.AuthUser)
    g.Get("profile", server.Controller(controller.UserController{}))
}
```

### 中间件

中间件类型为 `func(HandlerFunc) HandlerFunc`，按注册位置分为三层，执行顺序为：

```
Pre → 路由查找 → Use → 路由级中间件 → handler
```

- `app.Pre(...)`：在**路由查找之前**执行，可修改请求路径等影响路由匹配的行为（如去除 `.json` 后缀）
- `app.Use(...)`：在**路由查找之后**执行，可使用路由匹配结果（如访问日志、鉴权）
- 路由级：注册路由时的最后一个可变参数，仅对单个路由生效

```go
app.Pre(middleware.StripUrlSuffix)     // 路由查找前：去url后缀
app.Use(middleware.AccessLog)          // 路由查找后：访问日志
app.Get("/profile", h, middleware.User) // 路由级
```

已有的 `net/http` 生态组件可通过 `WrapHandler` / `WrapMiddleware` 接入：

```go
app.Get("/debug/pprof/*", server.WrapHandler(http.DefaultServeMux))
app.Use(server.WrapMiddleware(someHttpMiddleware))
```

### 静态文件

```go
// 将 url 前缀映射到目录：GET /static/* → ./assets/ 目录
app.Static("/static", "./assets")

// 单个文件
app.File("/favicon.ico", "./assets/favicon.ico")

// 嵌入资源（embed.FS）
//go:embed assets
var assets embed.FS
app.Get("/assets/*", server.AssetHandler("/assets/", assets, "assets"))
```

### 其他说明

- `app.Router().GetList()` 返回全部已注册的路由（方法+路径），可用于启动时打印路由表
- 404 默认由 `server.NotFoundHandler` 处理，可通过 `app.Get("/", server.NotFoundHandler)` 等方式自定义；统一错误/成功响应格式可覆盖 `app.HTTPErrorHandler` / `app.HTTPOKHandler`（见 examples 的 `framework/app.go`）
- 请求处理过程中的数据传递与响应输出见[上下文Context](#上下文context)一章（待补充）

#### WebSocket 与 Call（事件通道）

`WEBSOCKET` / `CALL` 方法路由与配套的事件机制用于**长连接 RPC** 场景：客户端（WebSocket 连接或进程内调用桥）发送的数据帧 `EventData{rid, uri, header, body}` 会被框架转换成内部请求——`uri` 决定路由、`body` 作为请求体、`rid` 用于请求响应配对——分发到 `app.WebSocket(uri, handler)` / `app.Call(uri, handler)` 注册的处理函数，处理结果按 `rid` 配对联回发送方。

连接侧的典型接入流程：

```go
// 1. 注册 WebSocket 方法的业务路由（数据帧的 uri 指向这里）
app.WebSocket("info", server.Controller(controller.InfoController{}))

// 2. 升级连接并挂到事件频道上（在控制器方法中完成）
func (ctl *CommonController) Connection(req *request.ConnectionRequest) error {
    conn, err := websocket.Accept(ctl.Ctx.ResponseWriter(), ctl.Ctx.Request(), opts)
    if err != nil {
        return err
    }
    subject := "market"                                    // 频道名，服务端推送时按它定位
    event := ctl.Ctx.Event(server.MethodWebSocket, subject)
    // 阻塞服务该连接：接收数据帧→路由分发→响应回送，连接关闭后返回
    return event.Serve(ctl.Ctx, ctl.Ctx.RequestID(), server.WebSocketWrapper(ctl.Ctx, conn))
}
```

服务端主动推送（如行情下发）按"方法+频道"广播：

```go
ctx.Event(server.MethodWebSocket, "market").Broadcast(ctx, &server.EventData{
    URI:  "quote/EURUSD",
    Body: body,
})
```

`Call` 方法同理，但连接来自进程内调用桥 `server.CallWrapper(ctx, callback)`，适合同进程模块间复用路由体系做 RPC。典型应用形态：交易/行情桥接服务接受客户端 WebSocket 订阅，业务请求走 `app.WebSocket(uri, controller)` 复用同一套控制器，服务端通过 `Broadcast` 向频道内的所有连接推送数据。

## 控制器

### 控制器的定义

控制器结构体必须包含类型为 Context 的 Ctx 成员变量。其定义如下所示：

```go
// UserController 定义一个User控制器
type UserController struct {
    Ctx Context
}
```

### 控制器方法的定义规则

控制器的 Public 类型成员函数用于处理业务逻辑，其入参和出参定义形式必须遵循以下规则。

- 入参必须有且仅有一个，且此参数必须实现 server.Request 接口。
- 返回参数可以有一个或两个，且最后一个返回参数必须是 error 类型。
- 当返回仅有1个 error 类型参数时，如果返回的 error 为 nil，框架将不会向HTTP输出任何内容。这可以方便开发者在控制器方法中自行定义HTTP响应的内容。如果返回 error 不为 nil，则根据 `framework/exception.go` 中定义的格式输出错误信息到HTTP响应。
- 当返回有2个参数时，如果返回的第二个参数 error 为 nil，则会根据 server.HTTPOKHandler 中定义的格式，输出第一个参数到HTTP响应。如果返回 error 不为 nil，处理方式与1个参数时相同，直接向HTTP响应输出错误信息。
- 控制器方法执行过程中发生 panic 时，框架会捕获并统一返回 500 错误响应（格式同 error 处理），不会导致连接中断。

```go
// 自定义响应内容的控制器方法
func (c *UserController) Login(req request.UserLoginRequest) error {
    c.Ctx.HTMLBlob(200, []byte("<h1>lazygo framework</h1>"))
    return nil
}

// 返回一个request.UserProfileResponse结构，框架默认会使用ctx.Succ()处理返回的*request.UserProfileResponse
// 也就是会将第一个参数*request.UserProfileResponse整个放入"data"字段中，并序列化成json。{"data": {"name": "李某人"}}
func (c *UserController) Profile(req request.UserProfileRequest) (*request.UserProfileResponse, error) {
    resp := &request.UserProfileResponse{
        Name: "李某人",
    }
    return resp, nil
}

// 返回一个map结构，返回结果处理方式同上 {"data": {"lazygo": "yes"}}
func (c *UserController) ProfileMap(req request.UserProfileRequest) (interface{}, error) {
    resp := map[string]interface{}{
        "lazygo": "yes",
    }
    return resp, nil
}

// 返回一个error信息，返回结果类似 {"errno": 400, "msg": "参数错误"}
// 可在 framework/exception.go 中自行定义返回结果的渲染格式
func (c *UserController) ProfileError(req request.UserProfileRequest) (interface{}, error) {
    return nil, server.NewHTTPError(200, 400, "参数错误")
}
```

需要注意的是，框架会默认认为所有 Public 成员函数都是HTTP请求处理函数，也就是说这些 Public 类型成员函数都必须遵循上述入参和出参的定义规则。在框架启动时会强制检查该控制器下所有的 Public 成员函数是否满足此规则。如果有不符合此规则的函数，请定义为开头小写的私有类型。

例如 `(ctl *UserController) Register(req *request.User) (any, error)` 和 `(ctl *UserController) Login(req *request.User) (any, error)` 两个函数内部都会调用发送通知的方法，可将发送通知定义为 `(ctl *UserController) sendMsg(uid uint64, msg string)` 并在 Register 和 Login 中以 `ctl.sendMsg(uid, msg)` 的形式调用。

### 注册控制器到HTTP路由

控制器函数需要注册到路由才能被HTTP请求访问到，注册路由时需要使用 `server.Controller` 函数将控制器函数转换为路由中的 HandlerFunc。

```go
// server.Controller 函数定义如下
// 参数h为控制器结构体实例
// method为函数名，若不指定函数名，会自动指定函数名为路由uri最后一个"/"后面的字符串
func Controller(h interface{}, methodName ...string) server.HandlerFunc
```

注册路由示例代码：

```go
// 将uri /user/profile 注册到UserController.UserInfo
app.Post("/user/profile", server.Controller(controller.UserController{}, "UserInfo"))

// 将uri /user/login 注册到UserController.Login
app.Post("/user/login", server.Controller(controller.UserController{}))
```

### Request 参数绑定和预处理

```go
type ToolsUploadRequest struct {
    Category string      `json:"category" bind:"query,form" process:"trim,cut(20)"`
    Tags     []int       `json:"tag" bind:"query,form"`
    Image    server.File `json:"image"`
}
```

在框架收到HTTP请求时，请求数据会自动绑定到 Request 结构体中。绑定需要依赖结构体注解来完成。

- 对于 `Content-Type` 为 `application/json` 的请求，会自动将json数据解析到结构体中。
- 对于 `Content-Type` 为 `form-data` 类型的请求或GET请求，可通过 `bind` 注解指定绑定的数据来源。例如 `bind:"query,form"` 表示优先从url的query参数中获取字段，如果获取不到，则使用form获取。
- bind 支持的类型为：
  - `value` 或 `ctx` 表示调用 context 中 Value 方法获取数据；
  - `header` 表示从HTTP Header中获取数据；
  - `param` 从路由参数中获取数据；
  - `query` 从URL Query中获取数据；
  - `form` 从Post Form中获取数据；
  - `file` 绑定文件数据。
- 数据类型为切片时，需要提供的参数格式为使用逗号分隔的字符串，或json数组字符串。例如 `?tags=1,2,3` 或 `?tags=[1,2,3]` 都可以绑定到 `[]int` 类型。
- 绑定支持常用值类型：字符串、各整型/无符号整型、浮点、布尔（`?flag=true`）及其切片；无法解析的值不会覆盖字段默认值。

在绑定数据时，会自动通过注解中的预处理器函数，对被绑定的参数进行预处理。多个预处理器之间使用逗号隔开。

- `trim`：剔除字符串两端空字符；
- `tolower`：字符串转为小写；
- `toupper`：字符串转为大写；
- `cut(num int)`：如果utf8字符串字符数量超过num个，则将字符串截断。

`Request` 需要实现 `func Verify() error` 和 `func Clear()` 两个函数。

- `Verify` 函数在参数绑定后执行，用于对请求参数内容进行校验。
- `Clear` 会在http请求返回响应后执行，用于做一些清理工作。

```go
func (r *ToolsUploadRequest) Verify() error {
    if !slices.Contains(utils.ImageFormat, path.Ext(r.Image.FileHeader.Filename)) {
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

## 上下文Context

> 该章节文档待补充。

## 日志

> 该章节文档待补充。

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

`db.Table()` 无需类型参数；写操作直接链式调用；读操作在终端方法（`Find` / `First`）上指定类型参数。每次返回一个新的查询构建器，构建器不是 goroutine 安全的，链式调用时暂存的条件、ORDER BY、LIMIT 等状态仅属于本次串式操作。

**查询**

```go
// Where 自动识别条件类型，可多次链式调用，多个条件之间以 AND 连接
users, _, err := db.Table("user").
    Where("name", "张三").                  // `name` = ?
    Where("age", ">", 18).                  // `age` > ?
    Where("status", "in", []any{1, 2}).     // `status` IN(?, ?)
    Where("deleted_at", nil).               // `deleted_at` IS NULL
    WhereMap(map[string]any{"age >=": 20}). // `age` >= ?
    OrderBy("id", "desc").
    Limit(10).
    Offset(20).
    Find[User]()
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
    Where("age", ">", 18).                  // 指定运算符：运算符原样拼入，如 > >= < <= != like contains 等
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
- key 含空格时按空格拆分为 `字段 运算符`，运算符支持 `>` `>=` `<` `<=` `!=` `in` `not in` `like` `contains`（模糊匹配语义见下文“模糊匹配”一节）
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

**模糊匹配（like / contains）**

- `like`：模式**原样**绑定为参数，由调用方自行处理 `%` 包裹与通配符转义，可表达前缀（`kw%`）、后缀（`%kw`）、多段等任意模式：

```go
cond["name like"] = sqldb.LikePattern(kw) // 安全模糊匹配：转义 %、_、\ 后包裹为 %kw%
cond["name like"] = "prefix%"             // 前缀匹配（自行构造模式）
```

- `contains`：固定"包含匹配"语义，生成 `LIKE ?`，参数自动经 `LikePattern` 转义并包裹为 `%kw%`。关键字中的 `%`、`_`、`\` 按字面量匹配，不会意外改变匹配行为：

```go
db.Table("user").Where("name", "contains", kw) // `name` LIKE ?，参数为转义后的 %kw%
cond["title contains"] = kw                    // WhereMap 形式，效果相同
```

`LikePattern` 的转义符为反斜杠，兼容 MySQL / PostgreSQL / SQLite 的默认 LIKE 语义；如需原生转义符行为（如开启 `NO_BACKSLASH_ESCAPES` 的 MySQL），请使用 `WhereRaw` 自行处理。

**排序、分组与分页**

```go
// 排序：方向仅支持 ASC / DESC（不区分大小写），可多次调用按顺序拼接
users, _, err := db.Table("user").Where("age", ">", 18).OrderBy("age", "desc").OrderBy("id", "asc").Find[User]()

// 随机排序
users, _, err = db.Table("user").OrderByRand().Find[User]()

// 指定查询字段：不调用Select时默认 SELECT *；Select 会按字段名排序后拼接
user, _, err := db.Table("user").Select("id", "name").Where("id", 1).First[User]()

// 分组：配合 Select 中的聚合函数做分组统计
type AgeStat struct {
    Age   int64 `json:"age"`
    Total int64 `json:"total"`
}
stats, _, err := db.Table("user").
    Select("age", "count(*) AS total").
    GroupBy("age").
    Find[AgeStat]()

// 分页：Limit 限制条数；Offset 为偏移量，需与 Limit 同时使用才生效
page, _, err := db.Table("user").Where("status", 1).OrderBy("id", "desc").Limit(10).Offset(20).Find[User]()
```

注意：不要对分组查询使用 `Count()`——`Count` 是单值查询，与 `GroupBy` 同用时不统计组数。分组统计请使用 `Select` + 聚合函数 + `Find` 的方式（如上面的 AgeStat 示例）。

**结果绑定**

- struct / []struct：按 json tag 匹配列名；匿名字段平铺；有 tag 的嵌套结构体以 tag 为前缀（如 `U.id`）；tag 带选项（`json:"uid,omitempty"`）也能正确匹配；`sql.NullString` 等 database/sql 类型直接作为叶子字段匹配
- `map[string]string` / `map[string]any`：值为字符串类型

**单条、单字段与统计**

```go
user, n, err := db.Table("user").Where("id", 1).First[User]()
// n == 0 表示记录不存在，err 为 nil，此时 user 为 User 零值

// 查询单个字段，支持表达式与别名；查不到时返回 ErrNoData
name, err := db.Table("user").Where("id", 1).One("name")
count, err := db.Table("user").Where("age", ">", 18).One("count(*) AS count")

total, err := db.Table("user").Where("age", ">", 18).Count()

// 判断记录是否存在
exists, err := db.Table("user").Where("id", 1).Exists()

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

防止误操作：`Update` / `UpdateRaw` / `Increment` / `Delete` 必须携带 WHERE 条件，否则返回 `ErrEmptyCond`。确实需要全表操作时，请显式传入 `Where("1=1")`。

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

list, _, err := db.Table(table.String()).
    Where("U.id", 1).
    Select(sqldb.StructFields(UserOrder{})...).
    Find[UserOrder]()
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
users, _, err := builder.Where("id", 1).Find[User]()
```

**构建器状态说明**

- 每个终端方法（`Find` / `First` / `Exists` / `One` / `Count` / `Insert` / `Update` / `UpdateRaw` / `Increment` / `Delete`）执行后会自动清空条件、ORDER BY、LIMIT 等暂存状态，builder 可继续复用
- `QueryString` 是纯拼接，不清理任何状态

**原生SQL**

```go
users, _, err := db.Find[User]("SELECT * FROM user WHERE age > ?", 18)

user, _, err := db.First[User]("SELECT * FROM user WHERE id = ?", 1)

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

// 查询：空结果返回零值, 0, nil，需同时判断 err 和 n
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

泛型缓存接口 `Cache[T]`，值统一以 JSON 序列化存储，支持 `memory`（进程内LRU）、`redis`、`memcache` 三种适配器。

#### 初始化

适配器的 `Option` 中，`name` 指向底层已初始化的实例名，`prefix` 为key前缀：

```go
// memory适配器依赖memory包
memory.Init([]memory.Config{
    {Name: "local", Capacity: 10000},
})
// redis/memcache适配器依赖对应包的实例
redis.Init([]redis.Config{
    {Name: "main", Host: "127.0.0.1", Port: 6379, Password: "123456", Db: 0},
})

err := cache.Init([]cache.Config{
    {Name: "default", Adapter: "memory", Option: map[string]string{"name": "local"}},
    {Name: "session", Adapter: "redis", Option: map[string]string{"name": "main", "prefix": "sess_"}},
}, "default")
```

#### 使用

每个命名实例都可以是强类型句柄，`Get`/`Remember` 直接返回类型化数据：

```go
// 强类型实例
sess, err := cache.Instance[SessionData]("session")

// 未命中返回 (零值, false, nil)
data, ok, err := sess.Get("abc")

// 缓存穿透：命中直接返回(hit=true)，未命中调用fn并写入缓存
data, hit, err := sess.Remember("abc", func() (SessionData, error) {
    return loadFromDB(), nil
}, 300)

err = sess.Set("abc", data, 300)  // ttl单位秒
exists, err := sess.Has("abc")
results, err := sess.HasMulti("a", "b", "c")  // map[string]bool
err = sess.Forget("abc")

// 默认实例的包级函数（同样支持泛型）
user, ok, err := cache.Get[User]("user_1")
user, hit, err := cache.Remember[User]("user_1", func() (User, error) {
    return loadUser(), nil
}, 300)
err = cache.Set("k", v, 60)
err = cache.Forget("k")
```

缓存值支持任意可 JSON 序列化的类型（结构体、切片、map、基本类型）。

### 分布式锁

基于 redis 的分布式锁（`SET NX` + 释放token校验），支持进程内串行、自旋等待与随机抖动。

#### 初始化

```go
err := locker.Init([]locker.Config{
    {
        Name:    "lock",
        Adapter: "redis",
        Option:  map[string]string{"name": "lazygo-rds"}, // 指向redis包的实例名
    },
}, "lock")
```

#### 使用

```go
// 自旋锁：拿不到锁时自旋等待，ctx控制最长等待时间
// ttl为锁的过期时间（秒），必须大于0，进程崩溃时锁靠ttl自动释放
rel, err := locker.Lock(ctx, "order:123", 30)
if err != nil {
    return err
}
defer rel.Release()
// ... 临界区操作

// 非阻塞尝试
rel, ok, err := locker.TryLock("order:123", 30)
if err != nil {
    return err
}
if ok {
    defer rel.Release()
}

// 带锁执行（默认实例，泛型版，fn的错误正常透传）
count, err := locker.LockFunc(ctx, "order:123", 30, func() (int, error) {
    return doSomething()
})
```

说明：

- `Lock` 的 `ctx` 是**获取阶段**的等待/取消语义，获取成功后不再受其过期影响
- `Release` 无参、幂等，内部使用带3秒超时的独立ctx执行释放，可重复调用
- 释放使用 Lua 脚本做 token 校验，只删除自己持有的锁，不会误删他人锁
- 同进程内相同资源串行竞争，避免对 redis 的无效压力

### 错误处理

> 该章节文档待补充。

## 动态库（c-shared）示例

lazygo 搭建的服务可编译为 C 共享库（`-buildmode=c-shared`），供 C 程序、脚本引擎（如 MQL）等
宿主加载：库内置的 HTTP/WS 服务器跑在 Go 自己的 goroutine 上，宿主通过
`Init`/`EventLoop`/`AsyncSend`/`Request` 四个 C API 与库双向交互；每个请求经三种
**投递模式**之一线程安全地交给宿主执行——函数指针回调（纯 C 宿主）、Windows 窗口消息
（GUI 宿主）、事件队列轮询（脚本/服务宿主，推荐）——ABI 支持宽字符（`wchar_t`）与
窄字符（`char`，UTF-8）两种构建，随 `UNICODE` 开关切换。代码位于 [examples/dylib](examples/dylib)。

### 宿主 API（hostbridge.h）

| 函数 | 方向 | 说明 |
|---|---|---|
| `Init(addr, port, dataDir, mode, handle)` | 宿主→库 | 启动服务；`mode` 取 `DELIVERY_MODE_*`，`handle` 随模式为回调函数指针或窗口句柄 |
| `EventLoop(uri, szURI, resp, szResp)` | 库→宿主 | 事件队列模式：宿主主循环轮询取请求（返回 rid，0 表示空） |
| `AsyncSend(rid, uri, header, body)` | 宿主→库 | 宿主回送/推送事件，按 rid 配对挂起的请求 |
| `Request(rid, uri, header, body, resp, szResp)` | 宿主→库 | 同步调用：派发给库内部 `/event/<uri>` 路由并返回其结果（不经过宿主） |

### 目录结构

```text
examples/dylib/
├── main.go          // //export 宿主 API 入口（Init/AsyncSend/Request/EventLoop）
├── bridge.go        // callHandle 投递分发 + TCHAR↔Go string 转换
├── server.go        // lazygo 服务启动 + echo 演示路由 + 事件桥（CallWrapper）
├── main.h/main.c    // 内部 C 头与助手（TCHAR 宏、InvokeCallback/InvokeSend、CopyString）
├── hostbridge.h     // 对外公共头（自包含，宿主侧只需它）
└── tests/dylib_test.c  // C 接口测试（EventLoop/CALLBACK 双模式、HTTP 全链路回环）
```

### 构建

依赖：Go 1.27+、CGO；交叉编译 Windows 需 mingw-w64（如 llvm-mingw）。

```bash
# macOS
CGO_ENABLED=1 go build -buildmode=c-shared -ldflags "-w -s" -trimpath -o libhostbridge.dylib .

# Windows（交叉编译示例，CC 指向 mingw clang）
CGO_ENABLED=1 CC=x86_64-w64-mingw32-clang GOOS=windows GOARCH=amd64 \
  go build -buildmode=c-shared -ldflags "-w -s -extldflags '-static'" -trimpath -o hostbridge.dll .

# Linux
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -buildmode=c-shared -ldflags "-w -s" -trimpath -o libhostbridge.so .
```

构建同时生成同名 `.h`（cgo 导出声明）。对外契约以 `hostbridge.h` 为准。

### 字符宽度（UNICODE 开关）

`main.h` 的 `#define UNICODE 1` 决定 ABI：宽字符（`wchar_t`，MQL 等宿主必须）
或窄字符（`char`，UTF-8，纯 C 宿主）。宿主包含 `hostbridge.h` 时须用 `-DUNICODE=`
保持与库的构建一致。

### 运行测试

```bash
# 构建 dylib 到 /tmp 后：
cc -I. tests/dylib_test.c -DUNICODE=1 -L/tmp -lhostbridge -o /tmp/dylib_test
DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test            # EVENT_LOOP 模式（生产同款）
DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test callback   # CALLBACK 模式（函数指针回调）
```
