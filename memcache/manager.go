package memcache

import (
	"errors"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/tidwall/gjson"
	"sync"
	"time"
)

type Manager struct {
	mc   map[string]*Memcache
	conf *gjson.Result
	lock *sync.RWMutex
}

/*
[
	{"name": "mc1", "server": [{"host": "127.0.0.1", "port": 11211}, {"host": "127.0.0.1", "port": 11212}]},
	{"name": "mc2", "server": [{"host": "127.0.0.1", "port": 11213}, {"host": "127.0.0.1", "port": 11214}]},
]
*/

func NewManager(conf *gjson.Result) (*Manager, error) {
	m := &Manager{
		mc:   map[string]*Memcache{},
		conf: conf,
		lock: new(sync.RWMutex),
	}
	for _, item := range conf.Array() {
		err := m.connect(&item)
		if err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *Manager) connect(item *gjson.Result) error {
	name := item.Get("name").String()
	server := item.Get("server").Array()

	var serverList []string
	for _, serv := range server {
		servMap := serv.Map()
		host := servMap["host"].String()
		port := servMap["port"].Int()
		serverList = append(serverList, fmt.Sprintf("%s:%d", host, port))
	}
	mc := memcache.New(serverList...)
	mc.MaxIdleConns = 10 // 最大保持10个空闲连接
	mc.Timeout = time.Duration(10) * time.Second

	m.lock.Lock()
	m.mc[name] = NewMemcache(name, mc)
	m.lock.Unlock()
	return nil
}

func (m *Manager) Mc(name string) (*Memcache, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	if mc, ok := m.mc[name]; ok {
		return mc, nil
	}
	return nil, errors.New("指定Memcached不存在")
}
