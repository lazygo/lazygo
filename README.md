# lazygo

lazygo框架


## 快速开始

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

或

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


项目文件目录结构

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
│   ├── logger.go                    // 提供日志记录器
│   └── route.go                     // 路由缓存
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

lazygo 默认支持toml和json两种格式的配置。

```
// 加载toml格式的配置
Toml(data []byte) (*Config, error)
// 加载json格式的配置
Json(data []byte) (*Config, error)

// 将配置信息中的field配置段解析到回调函数f的第一个参数中
// 因此需要确保提供的配置内容与回调函数的第一个参数的数据字段相匹配
func (c *Config) Register(field string, f interface{}) error
```

项目的配置加载相关代码在`config/config.go`中，可根据实际使用情况修改此文件内容。

示例：加载json配置
```
import (
    "github.com/lazygo/lazygo/config"
)


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
    }
    `

    // 使用json加载器解析配置内容
    jsonLoader, err := config.Json(data)
	if err != nil {
		return err
	}

	// 注册redis配置
	err = jsonLoader.Register("redis", func(conf []RedisConfig) error {
		fmt.Pringln(conf)
	})
	if err != nil {
		return err
	}

}

```

## 路由

## 控制器

控制器函数用于处理业务逻辑，控制器每一个函数都应通过路由绑定到指定的uri中。

控制器函数的参数为一个`Request`结构体

返回参数为一个`Response`结构体和一个error，如果返回error不为nil，将向http请求返回错误信息。

特殊情况下，控制器函数返回参数也可以仅有一个error，用于没有响应内容，只关注是否成功的http请求

注册路由时需要使用`framework.Controller`函数 将控制器函数转换为路由中的HandlerFunc

```
// 参数h为控制器结构体实例
// method为函数名，若不指定函数名，会自动指定函数名为路由uri最后一个“/”后面的字符串
func Controller(h interface{}, methodName ...string) server.HandlerFunc
```

示例代码：

```
type UserController struct{}

func (c *UserController)Login(request.UserLoginRequest) error {
    return nil
}
func (c *UserController)Profile(request.UserProfileRequest) (*request.UserProfileResponse, error) {
    resp := &request.UserProfileResponse{
        Name: "李某人",
    }
    return resp, nil
}

// 将uri /user/profile 注册到UserController.UserInfo
app.Post("/user/profile", framework.Controller(controller.UserController{}, "UserInfo"))

// 将uri /user/login 注册到UserController.Login
app.Post("/user/login", framework.Controller(controller.UserController{}))
```

参数绑定规则

- 请求数据会自动绑定到Request结构体中。绑定需要依赖结构体注解来完成。

- 对于 `Content-Type`为 `application/json` 的请求，会自动将json数据解析到结构体中。

- 对于 `Content-Type`为 `form-data`类型的请求或GET请求，可通过 `bind` 注解 指定绑定的数据来源。例如 bind:"query,form" 表示优先从url的query参数中获取字段，如果获取不到，则使用form获取。

- bind支持的类型为：`value` context中WithValue存储的数据，`header`HTTP Header，`param`参数路由，`query` URL Query，`form` Post Form，`file` 文件。

- 数据类型为切片时，需要提供的参数格式为 使用逗号分隔的字符串，或json数组字符串。例如`?tags=1,2,3` 或`?tags=[1,2,3]`都可以绑定到`[]int`类型
    
预处理器

- 将参数解析到绑定的字段前，会使用与处理器对参数进行预处理。

- 内置的预处理器包括 `trim` ：剔除字符串两端空字符，`cut(num int)` ：如果utf8字符串字符数量超过num个，则将字符串截断。

- 多个预处理器之间使用逗号隔开。

```
type ToolsUploadRequest struct {
    Category string      `json:"category" bind:"query,form" process:"trim,cut(20)"`
    Tags     []int       `json:"tag" bind:"query,form"`
    Image    server.File `json:"image"``
}
```

`Request`需要实现 `func Verify() error` 和 `func Clear()` 两个函数

`Verify`函数在参数绑定后执行，用于对请求参数内容进行校验

`Clear`会在http请求返回响应后执行，用于做一些清理工作

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