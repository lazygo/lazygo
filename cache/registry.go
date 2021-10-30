package cache

import "sync"

type adapter interface {
	init(map[string]string) (Cache, error)
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

