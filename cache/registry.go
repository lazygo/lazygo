package cache

import "sync"

type adapter interface {
	Cache
	init(map[string]string) error
	initialized() bool
}

type register struct {
	sync.Map
}

var registry = register{}

// add 注册适配器
func (r *register) add(name string, a adapter) {
	r.Store(name, a)
}

// get 获取适配器实例
func (r *register) get(name string) (adapter, error) {
	// 获取适配器
	if a, ok := r.Load(name); ok {
		return a.(adapter), nil
	}
	return nil, ErrAdapterNotFound
}

// init 初始化数据库连接
func (r *register) init(conf []*Config) error {
	for _, item := range conf {
		a, err := r.get(item.Name)
		if err != nil {
			return err
		}
		if a.initialized() {
			continue
		}
		err = a.init(item.Option)
		if err != nil {
			return err
		}
	}
	return nil
}