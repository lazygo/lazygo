package httpdns

import (
	"context"
	"net/netip"
	"sync"

	"github.com/lazygo/lazygo/internal"
)

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Option  map[string]string `json:"option" toml:"option"`
}

var registry = internal.Register[HTTPDNS]{}

type Manager struct {
	sync.Map
	defaultName string
}

var manager = &Manager{}

func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}

		a, err := registry.Get(item.Adapter)
		if err != nil {
			return err
		}
		lock, err := a.Init(item.Option)
		if err != nil {
			return err
		}
		m.Store(item.Name, lock)

		if defaultName == item.Name {
			m.defaultName = defaultName
		}
	}
	if m.defaultName == "" {
		return ErrInvalidDefaultName
	}
	return nil
}

type HTTPDNS interface {
	LookupIPAddr(context.Context, string) ([]netip.Addr, error)
}

// Init 初始化设置，在框架初始化时调用
func Init(conf []Config, defaultAdapter string) error {
	return manager.init(conf, defaultAdapter)
}

// Instance 获取分布式锁实例
func Instance(name string) (HTTPDNS, error) {
	a, ok := manager.Load(name)
	if !ok {
		return nil, ErrAdapterUninitialized
	}
	return a.(HTTPDNS), nil
}
