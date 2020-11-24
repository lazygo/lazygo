# lazygo

lazygo框架
### 功能简介


```
HTTP
  路由：自动路由
  中间件：不支持
  REST风格：不支持
  变量获取器：支持 json formdata
  错误处理：支持
HTTP请求：
  REST风格：支持
  文件上传：支持
  cookie：支持
  http代理：支持
配置文件：
  json格式配置文件: 支持
日志：
  console
  file
数据库
  mysql
    连接池：支持
    多数据库：支持
    ORM：不支持
    查询构建器：支持
    防注入：支持
  redis
    连接池：支持
    多数据库：支持
    使用redigo组件
  memcached：
    连接池：支持
    多数据库：支持
    分片：支持
  lruCache:
    本地内存缓存

模板：原生模板

定时任务：支持

一键打包资源文件：支持makefile

helper方法


```
### 生命周期

框架启动

```
初始化config
->初始化mysql、redis、memcached、cache、cron等组件
->注册路由、注册定时任务
->初始化资源
->常驻内存等待http请求
```

http请求-->http响应

```
匹配路由
通过反射创建控制器实例
读取request
业务代码
响应
```

命名规范：
 控制器: 需继承Controller，命名必须为XxxController
 
 控制器方法: 必须已Action结尾，否则自动路由加载不到
 
 model: 需继承Model
 
 定时任务Task: 需继承Task，命名必须为XxxTask
 
 定时任务方法 必须已Action结尾，否则加载不到

### 路由

控制器需要注册，注册之后，会按照请求路径匹配控制器和函数

```
// 自动路由
router.RegisterController(&controller.User{})
```

### 变量获取

```
ctl.GetString(name string) string
ctl.GetInt(name string) int
ctl.PostString(name string) string
ctl.PostInt(name string) int
ctl.GetHeader(name string) string
```

### http 客户端


```
// 实例：post 请求json数据
var res map[string]interface{}
err := libhttp.Post(checkUpdateApi).SetHeader(header).SetParams(params).ToJSON(&res)
```

### 读取配置文件

配置文件使用json格式

```
serverConfig, err := config.Config.GetSection("server")
utils.CheckFatal(err)
hostArray := serverConfig.Get("host").Array()
ip := serverConfig.Get("ip").String()
port := serverConfig.Get("port").Int()
```

### 日志


```
func Emergency(f interface{}, v ...interface{})
func Alert(f interface{}, v ...interface{})
func Critical(f interface{}, v ...interface{})
func Error(f interface{}, v ...interface{})
func Warn(f interface{}, v ...interface{})
func Notice(f interface{}, v ...interface{})
func Info(f interface{}, v ...interface{})
func Debug(f interface{}, v ...interface{})

```

### mysql


```
// 获取数据库
// dbname 是数据库配置中的name字段
database, err = mysql.Database("dbname")

// 执行SQL
database.Query(query string) (*sql.Rows, error)
database.Exec(query string) (sql.Result, error)
database.GetAll(query string) ([]map[string]interface{}, error)
database.GetAll(query string) ([]map[string]interface{}, error)

// 指定表名，获取查询构建器
builder := database.Table("table_name")

// 查询构建器使用

链式调用

// 设置条件

// 可以通过指定参数调用的方式设置条件，性能高。
builder.WhereMap(cond map[string]interface{})
.WhereClause(cond string)
.WhereIn(k string, in []interface{})
.WhereNotIn(k string, in []interface{})

或者通过自动匹配参数方式设置条件

// `k` IN('1', '2')
builder.Where("k", []interfece{}{1, 2})

// `name` = 'li' AND (`group`='admin' OR `group`='dev')
builder.Where(map[string]interface{}{"name": "li"}).Where("(`group`='admin' OR `group`='dev')")

// `name` = 'li' AND group IN("admin", "1")
cond := map[string]interface{}{
    "name": "li"
    "group": []interface{}{
        "admin",
        "dev",
    }
}
builder.Where(cond)


// 获取结果
builder.FetchOne(field string, order string, group string, start int) string
builder.FetchRow(fields string, order string, group string, start int) (map[string]interface{}, error)
builder.Fetch(fields string, order string, group string, limit int, start int) ([]map[string]interface{}, error)
builder.FetchWithPage(fields string, order string, group string, limit int, page int) (*ResultData, error)

// 聚合查询
builder.Count() int64

// 插入
builder.Insert(set map[string]interface{}) (int64, error)

// 更新
builder.Update(set map[string]interface{}, limit ...int) (int64, error)

// string column 自增的字段
// amount 自增数量
// column set 同时更新的字段（可选参数）
builder.Increment(column string, amount int, set ...map[string]interface{}) (int64, error)
builder.Decrement(column string, amount int, set ...map[string]interface{}) (int64, error)

// 删除
builder.Delete(limit ...int) (int64, error)

```

### redis、memcached 

方法太多，暂不做详细介绍

### 缓存

缓存支持memcached、redis和本地lru内存缓存，对于分布式系统主要使用memeceched或redis适配器

初始化

```
adapterName := "redis"
opt := map[string]interface{}{
    "name": "default", // db连接mame
}
cache.Init(adapterName, opt)
```

操作

```
// 设置缓存
cache.Remember(key string, value interface{}, timeout time.Duration) DataResult
// 获取缓存
cache.Get(key string) (DataResult, error)
// 判断缓存是否存在
cache.Has(key string) bool
// 删除缓存
cache.Forget(key string) bool

type DataResult interface {
	GetType() int32
	ToString() (string, error)
	ToByteArray() ([]byte, error)
	ToMap() (map[string]interface{}, error)
	ToMapArray() ([]map[string]interface{}, error)
}

// Remember的value参数可以是 DataResult 中的类型，或返回DataResult中类型的函数

callcack := func () interface {
    mdlUser := model.NewUser()
    cond := map[string]interface{}{
        "uid": 112,
    }
    data, err := mdlUser.GetInfo(cond)
    utils.CheckError(err)
    return data
}

res := cache.Remember("xxxxx_user_uid_112", callcack, 10*time.Minute)

data := res.ToMap()

```

### 分布式锁

初始化

```
adapterName := "redis"
opt := map[string]interface{}{
    "name": "default", // db连接mame
}
locker.Init(adapterName, opt)
```

使用示例1：

手动获取锁，释放锁

```
// 资源标识符
resource := "lock_resource_key_uid_1"

// 获取锁，并给锁设置10秒有效期
// 如果并发，等待200ms重试，共重试50次
// 如果需要立即返回，retry参数传0即可
lock, ok, err := locker.TryLock(resource, 10, 50)
if (err != nil) {
    // 错误处理
    panic(err)
}
if (!ok) {
    panic("并发了，获取锁超时")
}
// 获取到锁
// 处理业务逻辑
// 释放锁，建议使用defer 在处理业务逻辑之前
lock.Release()
```

使用示例2：
为函数增加方并发处理，自释放锁

```
// 资源标识符
resource := "lock_resource_key_uid_1"
// 设置防止并发的函数
f := func () interface{} {
    // 处理业务逻辑
    // 必须有返回值
    return "我是一个字符串"
}
// 获取锁，并给锁设置10秒有效期
// 重试次数*重试间隔时间自动控制，无需设置
// 函数执行完锁会自动释放
res, err := locker.LockFunc(resource, 10, f)

if (err != nil) {
    // 获取锁超时或其他错误，错误处理
    // 函数未被执行
    panic(err)
}

// 函数正常执行，获取结果
fmt.Println(res.(string))

```

### 模板


```
ctl.AssignMap(data map[string]interface{}) *Template

ctl.Assign(key string, data interface{}) *Template

ctl.Display(tpl string)
```

### 定时任务

分布式部署时 ，请注意定时任务可能会有问题，
如果只允许一台机器执行cron，可根据主机名来进行限制

```
RegisterTask(c, "*/5 * * * *", &task.DomainTask{}, "due")
RegisterTask(c, "*/6 * * * *", &task.DomainTask{}, "release")
RegisterTask(c, "*/1 * * * *", &task.DomainTask{}, "unlock")
```

### 工具方法

变量相关
ToString(value interface{}, defVal ...string) string
ToInt64(value interface{}, defVal ...int64) int64
ToInt(value interface{}, defVal ...int) int

字符串相关
ToCamelString(s string) string
VersionCompare(version1, version2, operator string) bool

文件相关
GetMimeType(name string) string
PathIsExist(f string) bool

日志相关

Debug(v ...interface{})
Info(v ...interface{})
Warn(v ...interface{})
Error(v ...interface{})
Fatal(v ...interface{})

错误相关
CheckError(err error)
CheckFatal(err error)

认证相关
CreateJwt(uid int64) string
VerifyJwt(tokenStr string) (int64, error)