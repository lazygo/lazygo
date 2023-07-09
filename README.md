# lazygo

lazygo框架


快速开始

```

# 安装lazygo
go install github.com/lazygo/lazygo@latest

# 创建项目
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

# 初始化项目
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