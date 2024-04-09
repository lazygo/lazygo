package internal

import (
	"errors"
	"sync"
)

var (
	ErrAdapterUndefined = errors.New("undefined adapter")
)

type adapter[T any] interface {
	Init(map[string]string) (T, error)
}

type adapterFunc[T any] func(map[string]string) (T, error)

func (fn adapterFunc[T]) Init(opt map[string]string) (T, error) {
	return fn(opt)
}

type Register[T any] struct {
	sync.Map
}

// add 注册适配器
func (r *Register[T]) Add(name string, f func(map[string]string) (T, error)) {
	r.Store(name, adapterFunc[T](f))
}

// get 获取适配器实例
func (r *Register[T]) Get(name string) (adapter[T], error) {
	// 获取适配器
	if a, ok := r.Load(name); ok {
		return a.(adapter[T]), nil
	}
	return nil, ErrAdapterUndefined
}
