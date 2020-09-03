package lazygo

import (
	"errors"
	"github.com/lazygo/lazygo/cache"
	"github.com/lazygo/lazygo/memcache"
	"github.com/lazygo/lazygo/mysql"
	"github.com/lazygo/lazygo/redis"
	"github.com/lazygo/lazygo/utils"
	"sync"
)

type Application struct {
	conf     *Config
	mysql    *mysql.Manager
	memcache *memcache.Manager
	redis    *redis.Manager
	cache    cache.Cache
	router   *Router
	asset    func(name string) ([]byte, error)
	server   *Server
	mu       *sync.RWMutex
}

type RouteRegister func(router *Router)
type AssetRegister func(name string) ([]byte, error)

var application *Application

func App() *Application {
	if application == nil {
		application = &Application{}
	}
	return application
}

func (a *Application) InitApp(configPath string, regRoute RouteRegister, regAsset AssetRegister) {
	config, err := NewConfig(configPath)
	utils.CheckFatal(err)
	a.conf = config

	// initMysql
	if conf, err := config.GetSection("mysql"); err == nil {
		a.mysql, err = mysql.NewManager(conf)
		utils.CheckFatal(err)
	}

	// initRedis
	if conf, err := config.GetSection("redis"); err == nil {
		a.redis, err = redis.NewManager(conf)
		utils.CheckFatal(err)
	}

	// initMemcache
	if conf, err := config.GetSection("memcached"); err == nil {
		a.memcache, err = memcache.NewManager(conf)
		utils.CheckFatal(err)
	}

	// initCache
	if conf, err := config.GetSection("cache"); err == nil {
		getAdapter := func(driver string, name string) interface{} {
			if driver == "redis"  {
				return a.GetRedis(name)
			}
			return a.GetMc(name)
		}
		a.cache, err = cache.NewCache(conf, getAdapter)
		utils.CheckFatal(err)
	}

	a.initRouter(regRoute)

	a.initAsset(regAsset)

	a.initServer()
}

func (a *Application) Run() {
	server := a.GetServer()
	server.Listen()
}

func (a *Application) GetDb(name string) *mysql.Db {
	if a.mysql == nil {
		panic(errors.New("mysql未初始化"))
	}
	db, err := a.mysql.Database(name)
	utils.CheckError(err)
	return db
}

func (a *Application) GetMc(name string) *memcache.Memcache {
	if a.memcache == nil {
		panic(errors.New("memcache未初始化"))
	}
	mc, err := a.memcache.Mc(name)
	utils.CheckError(err)
	return mc
}

func (a *Application) GetRedis(name string) *redis.Redis {
	if a.redis == nil {
		panic(errors.New("redis未初始化"))
	}
	pool, err := a.redis.RedisPool(name)
	utils.CheckError(err)
	return pool
}

func (a *Application) GetCache() cache.Cache {
	if a.cache == nil {
		panic(errors.New("缓存未初始化"))
	}

	return a.cache
}

func (a *Application) initRouter(route func(*Router)) bool {
	router := NewRouter()
	a.router = router
	route(a.router)
	return true
}

func (a *Application) initAsset(regAsset func(name string) ([]byte, error)) bool {
	a.asset = regAsset
	return true
}

func (a *Application) GetAsset(name string) ([]byte, error) {
	if a.asset == nil {
		panic(errors.New("资源未初始化"))
	}
	return a.asset(name)
}

func (a *Application) initServer() bool {
	if a.conf == nil {
		panic(errors.New("配置信息未初始化"))
	}
	if a.router == nil {
		panic(errors.New("路由未初始化"))
	}
	conf, err := a.conf.GetSection("server")
	utils.CheckError(err)

	server, err := NewServer(conf, a.router)
	utils.CheckError(err)
	a.server = server
	return true
}

func (a *Application) GetServer() *Server {
	if a.server == nil {
		panic(errors.New("服务器未初始化"))
	}
	return a.server
}
