package logs

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

// Legacy log level constants to ensure backwards compatibility.
const (
	LevelInfo  = LevelInformational
	LevelTrace = LevelDebug
	LevelWarn  = LevelWarning
)

// levelLogLogger is defined to implement log.Logger
// the real log level will be LevelEmergency
const levelLoggerImpl = -1

// Name for adapter with beego official support
const (
	AdapterConsole = "console"
	AdapterFile    = "file"
)

type newLoggerFunc func() Logger

// Logger defines the behavior of a log provider.
type Logger interface {
	Init(config string) error
	WriteMsg(when time.Time, msg map[string]interface{}, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]newLoggerFunc)
var levelPrefix = [LevelDebug + 1]string{"[M]", "[A]", "[C]", "[E]", "[W]", "[N]", "[I]", "[D]"}

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log newLoggerFunc) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, dup := adapters[name]; dup {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

// BeeLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type BeeLogger struct {
	lock                sync.Mutex
	level               int
	init                bool
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool
	prefix              string
	msgChanLen          int64
	msgChan             chan *logMsg
	signalChan          chan string
	wg                  sync.WaitGroup
	outputs             []*nameLogger
}

const defaultAsyncMsgLen = 1e3

type nameLogger struct {
	Logger
	name string
}

type logMsg struct {
	level int
	data  map[string]interface{}
	when  time.Time
}

var logMsgPool *sync.Pool

// NewLogger returns a new BeeLogger.
// channelLen means the number of messages in chan(used where asynchronous is true).
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channelLens ...int64) *BeeLogger {
	bl := new(BeeLogger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 2
	bl.msgChanLen = append(channelLens, 0)[0]
	if bl.msgChanLen <= 0 {
		bl.msgChanLen = defaultAsyncMsgLen
	}
	bl.signalChan = make(chan string, 1)
	bl.setLogger(AdapterConsole)
	return bl
}

// Async set the log to asynchronous and start the goroutine
func (bl *BeeLogger) Async(msgLen ...int64) *BeeLogger {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if bl.asynchronous {
		return bl
	}
	bl.asynchronous = true
	if len(msgLen) > 0 && msgLen[0] > 0 {
		bl.msgChanLen = msgLen[0]
	}
	bl.msgChan = make(chan *logMsg, bl.msgChanLen)
	logMsgPool = &sync.Pool{
		New: func() interface{} {
			return &logMsg{}
		},
	}
	bl.wg.Add(1)
	go bl.startLogger()
	return bl
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *BeeLogger) setLogger(adapterName string, configs ...string) error {
	config := append(configs, "{}")[0]
	for _, l := range bl.outputs {
		if l.name == adapterName {
			return fmt.Errorf("logs: duplicate adaptername %q (you have set this logger before)", adapterName)
		}
	}

	logAdapter, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}

	lg := logAdapter()
	err := lg.Init(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "logs.BeeLogger.SetLogger: "+err.Error())
		return err
	}
	bl.outputs = append(bl.outputs, &nameLogger{name: adapterName, Logger: lg})
	return nil
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
// config need to be correct JSON as string: {"interval":360}.
func (bl *BeeLogger) SetLogger(adapterName string, configs ...string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if !bl.init {
		bl.outputs = []*nameLogger{}
		bl.init = true
	}
	return bl.setLogger(adapterName, configs...)
}

// DelLogger remove a logger adapter in BeeLogger.
func (bl *BeeLogger) DelLogger(adapterName string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	outputs := []*nameLogger{}
	for _, lg := range bl.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			outputs = append(outputs, lg)
		}
	}
	if len(outputs) == len(bl.outputs) {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adapterName)
	}
	bl.outputs = outputs
	return nil
}

func (bl *BeeLogger) writeToLoggers(when time.Time, data map[string]interface{}, level int) {
	for _, l := range bl.outputs {
		err := l.WriteMsg(when, data, level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
		}
	}
}

func (bl *BeeLogger) writeMsg(logLevel int, data map[string]interface{}) error {
	if !bl.init {
		bl.lock.Lock()
		bl.setLogger(AdapterConsole)
		bl.lock.Unlock()
	}

	data["prefix"] = bl.prefix

	when := time.Now()
	if bl.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		data["filename"] = filename + ":" + strconv.Itoa(line)
	}

	//set level info in front of filename info
	if logLevel == levelLoggerImpl {
		// set to emergency to ensure all log will be print out correctly
		logLevel = LevelEmergency
	}
	data["levelFlag"] = levelPrefix[logLevel]

	if bl.asynchronous {
		lm := logMsgPool.Get().(*logMsg)
		lm.level = logLevel
		lm.data = data
		lm.when = when
		bl.msgChan <- lm
	} else {
		bl.writeToLoggers(when, data, logLevel)
	}
	return nil
}

// SetLevel Set log message level.
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (bl *BeeLogger) SetLevel(l int) {
	bl.level = l
}

// GetLevel Get Current log message level.
func (bl *BeeLogger) GetLevel() int {
	return bl.level
}

// SetLogFuncCallDepth set log funcCallDepth
func (bl *BeeLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// GetLogFuncCallDepth return log funcCallDepth for wrapper
func (bl *BeeLogger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// EnableFuncCallDepth enable log funcCallDepth
func (bl *BeeLogger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

// set prefix
func (bl *BeeLogger) SetPrefix(s string) {
	bl.prefix = s
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *BeeLogger) startLogger() {
	gameOver := false
	for {
		select {
		case bm := <-bl.msgChan:
			bl.writeToLoggers(bm.when, bm.data, bm.level)
			logMsgPool.Put(bm)
		case sg := <-bl.signalChan:
			// Now should only send "flush" or "close" to bl.signalChan
			bl.flush()
			if sg == "close" {
				for _, l := range bl.outputs {
					l.Destroy()
				}
				bl.outputs = nil
				gameOver = true
			}
			bl.wg.Done()
		}
		if gameOver {
			break
		}
	}
}

// Emergency Log EMERGENCY level message.
func (bl *BeeLogger) Emergency(data map[string]interface{}) {
	if LevelEmergency > bl.level {
		return
	}
	bl.writeMsg(LevelEmergency, data)
}

// Alert Log ALERT level message.
func (bl *BeeLogger) Alert(data map[string]interface{}) {
	if LevelAlert > bl.level {
		return
	}
	bl.writeMsg(LevelAlert, data)
}

// Critical Log CRITICAL level message.
func (bl *BeeLogger) Critical(data map[string]interface{}) {
	if LevelCritical > bl.level {
		return
	}
	bl.writeMsg(LevelCritical, data)
}

// Error Log ERROR level message.
func (bl *BeeLogger) Error(data map[string]interface{}) {
	if LevelError > bl.level {
		return
	}
	bl.writeMsg(LevelError, data)
}

// Warning Log WARNING level message.
func (bl *BeeLogger) Warning(data map[string]interface{}) {
	if LevelWarn > bl.level {
		return
	}
	bl.writeMsg(LevelWarn, data)
}

// Notice Log NOTICE level message.
func (bl *BeeLogger) Notice(data map[string]interface{}) {
	if LevelNotice > bl.level {
		return
	}
	bl.writeMsg(LevelNotice, data)
}

// Informational Log INFORMATIONAL level message.
func (bl *BeeLogger) Informational(data map[string]interface{}) {
	if LevelInfo > bl.level {
		return
	}
	bl.writeMsg(LevelInfo, data)
}

// Debug Log DEBUG level message.
func (bl *BeeLogger) Debug(data map[string]interface{}) {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(LevelDebug, data)
}

// Warn Log WARN level message.
// compatibility alias for Warning()
func (bl *BeeLogger) Warn(data map[string]interface{}) {
	if LevelWarn > bl.level {
		return
	}
	bl.writeMsg(LevelWarn, data)
}

// Info Log INFO level message.
// compatibility alias for Informational()
func (bl *BeeLogger) Info(data map[string]interface{}) {
	if LevelInfo > bl.level {
		return
	}
	bl.writeMsg(LevelInfo, data)
}

// Trace Log TRACE level message.
// compatibility alias for Debug()
func (bl *BeeLogger) Trace(data map[string]interface{}) {
	if LevelDebug > bl.level {
		return
	}
	bl.writeMsg(LevelDebug, data)
}

// Flush flush all chan data.
func (bl *BeeLogger) Flush() {
	if bl.asynchronous {
		bl.signalChan <- "flush"
		bl.wg.Wait()
		bl.wg.Add(1)
		return
	}
	bl.flush()
}

// Close close logger, flush all chan data and destroy all adapters in BeeLogger.
func (bl *BeeLogger) Close() {
	if bl.asynchronous {
		bl.signalChan <- "close"
		bl.wg.Wait()
		close(bl.msgChan)
	} else {
		bl.flush()
		for _, l := range bl.outputs {
			l.Destroy()
		}
		bl.outputs = nil
	}
	close(bl.signalChan)
}

// Reset close all outputs, and set bl.outputs to nil
func (bl *BeeLogger) Reset() {
	bl.Flush()
	for _, l := range bl.outputs {
		l.Destroy()
	}
	bl.outputs = nil
}

func (bl *BeeLogger) flush() {
	if bl.asynchronous {
		for {
			if len(bl.msgChan) > 0 {
				bm := <-bl.msgChan
				bl.writeToLoggers(bm.when, bm.data, bm.level)
				logMsgPool.Put(bm)
				continue
			}
			break
		}
	}
	for _, l := range bl.outputs {
		l.Flush()
	}
}
