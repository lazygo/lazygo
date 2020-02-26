package redis

import (
	"errors"
	"fmt"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/tidwall/gjson"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Manager struct {
	redis map[string]*Redis
	conf  *gjson.Result
	lock  *sync.RWMutex
}

/*
[
	{"name": "redis1", "host": "127.0.0.1", "port": 11211, "password": "secret", "db": 0},
	{"name": "redis2", "host": "127.0.0.1", "port": 11212, "password": "secret", "db": 0}
]
*/

func NewManager(conf *gjson.Result) (*Manager, error) {
	m := &Manager{
		redis: map[string]*Redis{},
		conf:  conf,
		lock:  new(sync.RWMutex),
	}
	for _, item := range conf.Array() {
		err := m.connect(&item)
		if err != nil {
			return nil, err
		}

	}
	m.closePool()
	return m, nil
}

func (m *Manager) connect(item *gjson.Result) error {

	name := item.Get("name").String()
	host := item.Get("host").String()
	port := item.Get("port").Int()
	password := item.Get("password").String()
	db := item.Get("db").Int()

	addr := fmt.Sprintf("%s:%d", host, port)

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

	if password != "" {
		dialOpts = append(dialOpts, redigo.DialPassword(password)) // 鉴权密码，默认空
	}
	if db != 0 {
		dialOpts = append(dialOpts, redigo.DialDatabase(int(db))) // 数据库号，默认0
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

	m.lock.Lock()
	m.redis[name] = NewRedis(name, pool)
	m.lock.Unlock()

	return nil
}

func (m *Manager) RedisPool(name string) (*Redis, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if redis, ok := m.redis[name]; ok {
		return redis, nil
	}
	return nil, errors.New("指定Redis不存在")
}

// closePool 程序进程退出时关闭连接池
func (m *Manager) closePool() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, syscall.SIGKILL)
	go func() {
		<-ch
		for _, redis := range m.redis {
			redis.pool.Close()
		}
		os.Exit(0)
	}()
}
