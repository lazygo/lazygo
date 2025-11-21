package logger

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/lazygo/lazygo/internal"
)

// RFC5424 log message levels.
const (
	LevelEmergency     = iota // 0
	LevelAlert                // 1
	LevelCritical             // 2
	LevelError                // 3
	LevelWarning              // 4
	LevelNotice               // 5
	LevelInformational        // 6
	LevelDebug                // 7
)

// Legacy log level constants to ensure backwards compatibility.
const (
	LevelWarn  = LevelWarning       // 4
	LevelInfo  = LevelInformational // 6
	LevelTrace = LevelDebug         // 7
)

var levelPrefix = [LevelDebug + 1]string{"[EMERGENCY]", "[ALERT]", "[CRITICAL]", "[ERROR]", "[WARNING]", "[NOTICE]", "[INFO]", "[DEBUG]"}

type Config struct {
	Name      string            `json:"name" toml:"name"`
	Adapter   string            `json:"adapter" toml:"adapter"`
	Async     bool              `json:"async" toml:"async"`
	Level     uint8             `json:"level" toml:"level"`
	Caller    bool              `json:"caller" toml:"caller"`
	CallDepth int               `json:"call_depth" toml:"call_depth"`
	Option    map[string]string `json:"option" toml:"option"`
}

var registry = internal.Register[logWriter, map[string]string]{}

type logWriter interface {
	Write([]byte, time.Time, string) (int, error)
	Close() error
	Flush() error
}

type Writer interface {
	io.WriteCloser

	Println(v ...any)

	// Emergency Log EMERGENCY level message.
	Emergency(v ...any)

	// Alert Log ALERT level message.
	Alert(v ...any)

	// Critical Log CRITICAL level message.
	Critical(v ...any)

	// Error Log ERROR level message.
	Error(v ...any)

	// Warning Log WARNING level message.
	Warning(v ...any)

	// Notice Log NOTICE level message.
	Notice(v ...any)

	// Informational Log INFORMATIONAL level message.
	Informational(v ...any)

	// Debug Log DEBUG level message.
	Debug(v ...any)

	// Warn Log WARN level message.
	// compatibility alias for Warning()
	Warn(v ...any)

	// Info Log INFO level message.
	// compatibility alias for Informational()
	Info(v ...any)

	// Trace Log TRACE level message.
	// compatibility alias for Debug()
	Trace(v ...any)
}

type writer struct {
	async     *asyncWriter
	lw        logWriter
	level     uint8
	caller    bool
	short     bool
	callDepth int
}

func newWriter(lw logWriter, config Config) Writer {
	w := &writer{
		lw:        lw,
		level:     config.Level,
		caller:    config.Caller,
		callDepth: config.CallDepth,
	}
	if config.Async {
		w.async = newAsync(lw, 1000)
	}
	return w
}

func (w *writer) write(b []byte, prefix string, callDepth int) (int, error) {
	t := time.Now()

	if w.caller {
		_, file, line, ok := runtime.Caller(callDepth + 1)
		if !ok {
			file = "???"
			line = 0
		}
		if w.short {
			_, file = path.Split(file)
		}
		file += ":" + strconv.Itoa(line) + " "
		b = append([]byte(file), b...)
	}
	if w.async != nil {
		return w.async.Write(b, t, prefix)
	}
	return w.lw.Write(b, t, prefix)
}

func (w *writer) Write(b []byte) (int, error) {
	return w.write(b, "[info]", w.callDepth+1)
}

func (w *writer) Close() error {
	if w.async != nil {
		w.async.Close()
		return nil
	}
	return w.lw.Close()
}

func (w *writer) Println(v ...any) {
	w.write(str2bytes(fmt.Sprintln(v...)), "[info]", w.callDepth+1)
}

// Emergency Log EMERGENCY level message.
func (w *writer) Emergency(v ...any) {
	if LevelEmergency > w.level {
		return
	}
	prefix := levelPrefix[LevelEmergency]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Alert Log ALERT level message.
func (w *writer) Alert(v ...any) {
	if LevelAlert > w.level {
		return
	}
	prefix := levelPrefix[LevelAlert]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Critical Log CRITICAL level message.
func (w *writer) Critical(v ...any) {
	if LevelCritical > w.level {
		return
	}
	prefix := levelPrefix[LevelCritical]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Error Log ERROR level message.
func (w *writer) Error(v ...any) {
	if LevelError > w.level {
		return
	}
	prefix := levelPrefix[LevelError]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Warning Log WARNING level message.
func (w *writer) Warning(v ...any) {
	if LevelWarn > w.level {
		return
	}
	prefix := levelPrefix[LevelWarn]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Notice Log NOTICE level message.
func (w *writer) Notice(v ...any) {
	if LevelNotice > w.level {
		return
	}
	prefix := levelPrefix[LevelNotice]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Informational Log INFORMATIONAL level message.
func (w *writer) Informational(v ...any) {
	if LevelInfo > w.level {
		return
	}
	prefix := levelPrefix[LevelInfo]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Debug Log DEBUG level message.
func (w *writer) Debug(v ...any) {
	if LevelDebug > w.level {
		return
	}
	prefix := levelPrefix[LevelDebug]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (w *writer) Warn(v ...any) {
	if LevelWarn > w.level {
		return
	}
	prefix := levelPrefix[LevelWarn]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (w *writer) Info(v ...any) {
	if LevelInfo > w.level {
		return
	}
	prefix := levelPrefix[LevelInfo]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (w *writer) Trace(v ...any) {
	if LevelDebug > w.level {
		return
	}
	prefix := levelPrefix[LevelDebug]
	w.write(str2bytes(fmt.Sprintln(v...)), prefix, w.callDepth+1)
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

		a, err := registry.Get(item.Adapter)
		if err != nil {
			return err
		}
		lw, err := a.Init(item.Option)
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
