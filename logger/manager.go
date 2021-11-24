package logger

import (
	"io"
	"sync"
	"time"
)

type Config struct {
	Name    string            `json:"name" toml:"name"`
	Adapter string            `json:"adapter" toml:"adapter"`
	Async   bool              `json:"async" toml:"async"`
	Level   uint8             `json:"level" toml:"level"`
	Option  map[string]string `json:"option" toml:"option"`
}

type logWriter interface {
	Writeln([]byte, time.Time) (int, error)
	Close() error
	Flush() error
}

type Writer interface {
	io.WriteCloser
}

type writer struct {
	async *asyncWriter
	lw    logWriter
	level uint8
}

func newWriter(lw logWriter, config Config) Writer {
	w := &writer{
		lw:    lw,
		level: config.Level,
	}
	if config.Async {
		w.async = newAsync(lw, 1000)
	}
	return w
}

func (w *writer) Write(b []byte) (int, error) {
	t := time.Now()

	if w.async != nil {
		return w.async.Writeln(b, t)
	}
	return w.lw.Writeln(b, t)
}

func (w *writer) Close() error {
	if w.async != nil {
		w.async.Close()
		return nil
	}
	return w.lw.Close()
}

type Manager struct {
	sync.Map
	defaultName string
}

var manager = &Manager{}

// init 初始化日志记录器
func (m *Manager) init(conf []Config, defaultName string) error {
	for _, item := range conf {
		if _, ok := m.Load(item.Name); ok {
			continue
		}

		a, err := registry.get(item.Adapter)
		if err != nil {
			return err
		}
		lw, err := a.init(item.Option)
		if err != nil {
			return err
		}
		m.Store(item.Name, newWriter(lw, item))

		if defaultName == item.Name {
			m.defaultName = defaultName
		}
	}
	if m.defaultName == "" {
		return ErrInvalidDefaultName
	}
	return nil
}

// Init 初始化设置，在框架初始化时调用
func Init(conf []Config, defaultAdapter string) error {
	return manager.init(conf, defaultAdapter)
}

// Instance 获取日志记录器实例
func Instance(name string) (Writer, error) {
	a, ok := manager.Load(name)
	if !ok {
		return nil, ErrAdapterUninitialized
	}
	return a.(Writer), nil
}
