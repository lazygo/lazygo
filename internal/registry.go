package internal

import (
	"errors"
	"sync"
)

var (
	ErrAdapterUndefined = errors.New("undefined adapter")
)

type Adapter[T any, O any] interface {
	Init(O) (T, error)
}

type adapterFunc[T any, O any] func(O) (T, error)

func (fn adapterFunc[T, O]) Init(opt O) (T, error) {
	return fn(opt)
}

type Register[T any, O any] struct {
	sync.Map
}

// Add 注册适配器
func (r *Register[T, O]) Add(name string, f func(O) (T, error)) {
	r.Store(name, adapterFunc[T, O](f))
}

// Get 获取适配器实例
func (r *Register[T, O]) Get(name string) (Adapter[T, O], error) {
	// 获取适配器
	if a, ok := r.Load(name); ok {
		return a.(Adapter[T, O]), nil
	}
	return nil, ErrAdapterUndefined
}
