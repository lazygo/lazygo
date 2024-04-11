# lazygo

lazygo框架


## 快速开始


### 初始化lazygo项目

方案一：
```

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

```

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

```
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

```
// 加载toml格式的配置
loader := config.Toml(data []byte) (*Config, error)
// 加载json格式的配置
loader := config.Json(data []byte) (*Config, error)
```

可以调用func (c *Config) Load(field string f any) 方法，将配置信息中的field配置段解析到回调函数f的第一个参数中。
其中回调函数f的定义需符合 func(*CustomStruct) error 或 func([]CustomStruct) error 的形式。 注意需要确保提供的配置内容与回调函数参数结构体类型CustomStruct的数据字段相匹配。如果加载的是json类型的配置，则需要在CustomStruct结构体字段中提供json注解，同理toml格式配置需要提供toml注解。
以下代码为解析配置redis配置段到RedisConfig结构体的示例。
```

示例：加载json配置
```
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

```
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

```
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


```
// server.Controller 函数定义如下
// 参数h为控制器结构体实例
// method为函数名，若不指定函数名，会自动指定函数名为路由uri最后一个“/”后面的字符串
func Controller(h interface{}, methodName ...string) server.HandlerFunc
```

注册路由示例代码：

```
// 将uri /user/profile 注册到UserController.UserInfo
app.Post("/user/profile", server.Controller(controller.UserController{}, "UserInfo"))

// 将uri /user/login 注册到UserController.Login
app.Post("/user/login", server.Controller(controller.UserController{}))
```

### Request 参数绑定和预处理 ###

```
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


```
func (r *ToolsUploadRequest) Verify() error {

	if utils.InSlice(utils.ImageFormat, path.Ext(r.Image.FileHeader.Filename)) == false {
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

查询构建器

### 缓存
