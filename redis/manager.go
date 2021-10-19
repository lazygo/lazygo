package redis

import (
	"fmt"
	redigo "github.com/gomodule/redigo/redis"
	"runtime"
	"sync"
	"time"
)

type Config struct {
	Name     string `json:"name" toml:"name"`
	Host     string `json:"host" toml:"host"`
	Port     int    `json:"port" toml:"port"`
	Password string `json:"password" toml:"password"`
	Db       int    `json:"db" toml:"db"`
	Prefix   string `json:"prefix" toml:"prefix"`
}

type Manager struct {
	sync.Map
}

// 单例
var manager = &Manager{}

// init 初始化数据库连接
func (m *Manager) init(conf []*Config) error {
	for _, item := range conf {
		if _, ok := manager.Load(item.Name); ok {
			// 已连接的就不再次连接了
			continue
		}
		pool, err := manager.open(item)
		if err != nil {
			return err
		}
		m.Store(item.Name, newRedis(item.Name, pool, item.Prefix))
	}
	return nil
}

// closeAll 关闭数据库连接
func (m *Manager) closeAll() error {
	var err error
	m.Range(func(name, db interface{}) bool {
		err = db.(*Redis).Close()
		if err != nil {
			return false
		}
		m.Delete(name)
		return true
	})
	return err
}

// 连接redis
func (m *Manager) open(item *Config) (*redigo.Pool, error) {

	addr := fmt.Sprintf("%s:%d", item.Host, item.Port)

	dialOpts := []redigo.DialOption{
		redigo.DialConnectTimeout(time.Millisecond * 500), // 连接超时，默认500*time.Millisecond
		redigo.DialReadTimeout(time.Second),               // 读取超时，默认time.Second
		redigo.DialWriteTimeout(time.Second),              // 写入超时，默认time.Second
		redigo.DialKeepAlive(time.Minute * 5),             // 默认5*time.Minute
		redigo.DialNetDial(nil),                           // 自定义dial，默认nil
		redigo.DialUseTLS(false),                          // 是否用TLS，默认false
		redigo.DialTLSSkipVerify(false),                   // 服务器证书校验，默认false
		redigo.DialTLSConfig(nil),                         // 默认nil，详见tls.Config
	}

	if item.Password != "" {
		dialOpts = append(dialOpts, redigo.DialPassword(item.Password)) // 鉴权密码，默认空
	}
	if item.Db != 0 {
		dialOpts = append(dialOpts, redigo.DialDatabase(item.Db)) // 数据库号，默认0
	}

	pool := &redigo.Pool{
		MaxIdle:         2 * runtime.GOMAXPROCS(0), // 最大空闲连接数，即会有这么多个连接提前等待着，但过了超时时间也会关闭
		MaxActive:       5000,                      // 最大连接数，即最多的tcp连接数，一般建议往大的配置，但不要超过操作系统文件句柄个数（centos下可以ulimit -n查看）
		IdleTimeout:     180 * time.Second,         // 空闲连接超时时间，应该设置比redis服务器超时时间短。否则服务端超时了，客户端保持着连接也没用
		Wait:            true,                      // 如果超过最大连接，是报错，还是等待
		MaxConnLifetime: 0,                         // 连接的生命周期，默认0不失效
		TestOnBorrow:    nil,                       // 空间连接取出后检测是否健康，默认nil
		Dial: func() (conn redigo.Conn, e error) {
			return redigo.Dial("tcp", addr, dialOpts...)
		},
	}

	return pool, nil
}

// Init 初始化数据库
func Init(conf []*Config) error {
	return manager.init(conf)
}

// CloseAll 关闭数据库连接
func CloseAll() error {
	return manager.closeAll()
}

// Pool 通过名称获取Redis
func Pool(name string) (*Redis, error) {
	if redis, ok := manager.Load(name); ok {
		return redis.(*Redis), nil
	}
	return nil, ErrRedisNotExists
}
