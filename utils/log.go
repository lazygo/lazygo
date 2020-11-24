package utils

import (
	"github.com/lazygo/lazygo/logs"
)

// beeLogger references the used application logger.
var beeLogger = logs.NewLogger()

// Reset will remove all the adapter
func Reset() {
	beeLogger.Reset()
}

// Async set the beelogger with Async mode and hold msglen messages
func Async(msgLen ...int64) *logs.BeeLogger {
	return beeLogger.Async(msgLen...)
}

// SetLevel sets the global log level used by the simple logger.
func SetLogLevel(l int) {
	beeLogger.SetLevel(l)
}

// SetPrefix sets the prefix
func SetLogPrefix(s string) {
	beeLogger.SetPrefix(s)
}

// SetLogFuncCall set the CallDepth, default is 4
func SetLogFuncCall(b bool) {
	beeLogger.EnableFuncCallDepth(b)
	beeLogger.SetLogFuncCallDepth(4)
}

// 设置日志
// adapter console、file、multifile
// beego.SetLogger("file", `{"filename":"logs/test.log"}`)
func SetLogger(adapter string, config ...string) error {
	return beeLogger.SetLogger(adapter, config...)
}

// Emergency logs a message at emergency level.
func Emergency(msg string, v ...map[string]interface{}) {
	beeLogger.Emergency(formatLog(msg, v...))
}

// Alert logs a message at alert level.
func Alert(msg string, v ...map[string]interface{}) {
	beeLogger.Alert(formatLog(msg, v...))
}

// Critical logs a message at critical level.
func Critical(msg string, v ...map[string]interface{}) {
	beeLogger.Critical(formatLog(msg, v...))
}

// Error logs a message at error level.
func Error(msg string, v ...map[string]interface{}) {
	beeLogger.Error(formatLog(msg, v...))
}

// Warn compatibility alias for Warning()
func Warn(msg string, v ...map[string]interface{}) {
	beeLogger.Warn(formatLog(msg, v...))
}

// Notice logs a message at notice level.
func Notice(msg string, v ...map[string]interface{}) {
	beeLogger.Notice(formatLog(msg, v...))
}

// Info compatibility alias for Warning()
func Info(msg string, v ...map[string]interface{}) {
	beeLogger.Info(formatLog(msg, v...))
}

// Debug logs a message at debug level.
func Debug(msg string, v ...map[string]interface{}) {
	beeLogger.Debug(formatLog(msg, v...))
}

// Trace logs a message at trace level.
// compatibility alias for Warning()
func Trace(msg string, v ...map[string]interface{}) {
	beeLogger.Trace(formatLog(msg, v...))
}

func formatLog(msg string, v ...map[string]interface{}) map[string]interface{} {
	data := map[string]interface{}{}
	if len(v) > 0 {
		data = v[0]
		data["ext_data"] = v[1:]
	}
	data["message"] = msg
	return data
}
