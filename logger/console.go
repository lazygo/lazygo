package logger

import (
	"io"
	"os"
	"sync"
	"time"
)

type consoleLogWriter struct {
	sync.Mutex
	writer io.Writer
}

func newConsoleLogWriter(opt map[string]string) (logWriter, error) {
	cl := &consoleLogWriter{
		writer: os.Stdout,
	}
	return cl, nil
}

func (cl *consoleLogWriter) Write(b []byte, t time.Time, prefix string) (int, error) {
	cl.Lock()
	defer cl.Unlock()
	hd, _, _ := formatTimeHeader(t)
	b = append(hd, b...)
	if prefix != "" {
		b = append([]byte(prefix+" "), b...)
	}
	return cl.writer.Write(b)
}

func (cl *consoleLogWriter) Close() error {
	return nil
}
func (cl *consoleLogWriter) Flush() error {
	return nil
}

func init() {
	// 注册适配器
	registry.add("console", adapterFunc(newConsoleLogWriter))
}
