package logs

import (
	"encoding/json"
	"fmt"
	"github.com/shiena/ansicolor"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// brush is a color join function
type brush func(string) string

// newBrush return a fix color Brush
func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

var colors = []brush{
	newBrush("1;37"), // Emergency          white
	newBrush("1;36"), // Alert              cyan
	newBrush("1;35"), // Critical           magenta
	newBrush("1;31"), // Error              red
	newBrush("1;33"), // Warning            yellow
	newBrush("1;32"), // Notice             green
	newBrush("1;34"), // Informational      blue
	newBrush("1;44"), // Debug              Background blue
}

type consoleLogWriter struct {
	sync.Mutex
	writer io.Writer
}

func newLogWriter(wr io.Writer) *consoleLogWriter {
	return &consoleLogWriter{writer: wr}
}

func (lg *consoleLogWriter) writeln(when time.Time, msg string) {
	lg.Lock()
	h, _, _ := formatTimeHeader(when)
	lg.writer.Write(append(append(h, msg...), '\n'))
	lg.Unlock()
}

// consoleWriter implements LoggerInterface and writes messages to terminal.
type consoleWriter struct {
	lg       *consoleLogWriter
	Level    int  `json:"level"`
	Colorful bool `json:"color"` //this filed is useful only when system's terminal supports color
}

// NewConsole create ConsoleWriter returning as LoggerInterface.
func NewConsole() Logger {
	cw := &consoleWriter{
		lg:       newLogWriter(ansicolor.NewAnsiColorWriter(os.Stdout)),
		Level:    LevelDebug,
		Colorful: true,
	}
	return cw
}

// 初始化console适配器
// jsonConfig like '{"level":LevelTrace}'.
func (c *consoleWriter) Init(jsonConfig string) error {
	if len(jsonConfig) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(jsonConfig), c)
}

// WriteMsg write message in console.
func (c *consoleWriter) WriteMsg(when time.Time, data map[string]interface{}, level int) error {
	if level > c.Level {
		return nil
	}
	msg := fmt.Sprint(data)
	if c.Colorful {
		msg = strings.Replace(msg, levelPrefix[level], colors[level](levelPrefix[level]), 1)
	}
	c.lg.writeln(when, msg)
	return nil
}

func (c *consoleWriter) Destroy() {}
func (c *consoleWriter) Flush()   {}

func init() {
	Register(AdapterConsole, NewConsole)
}
