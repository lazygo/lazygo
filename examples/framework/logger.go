package framework

import (
	"fmt"
	"log"
	"strings"

	"github.com/lazygo/lazygo/logger"
)

type Logger interface {
	Alert(format string, a ...interface{})  // 1
	Error(format string, a ...interface{})  // 3
	Warn(format string, a ...interface{})   // 4
	Notice(format string, a ...interface{}) // 5
	Info(format string, a ...interface{})   // 6
	Debug(format string, a ...interface{})  // 7
}

type loggerImpl struct {
	Ctx Context `json:"context"`
}

var (
	AccessLog  logger.Writer
	ConsoleLog logger.Writer
	AppLog     logger.Writer
	ErrorLog   logger.Writer
)

func InitLogger() {
	var err error
	ErrorLog, err = logger.Instance("error-log")
	if err != nil {
		log.Fatalln("load error logger fail")
	}
	AccessLog, err = logger.Instance("access-log")
	if err != nil {
		log.Fatalln("load access logger fail")
	}
	ConsoleLog, err = logger.Instance("console-log")
	if err != nil {
		log.Fatalln("load console logger fail")
	}
	AppLog, err = logger.Instance("app-log")
	if err != nil {
		log.Fatalln("load app logger fail")
	}
}

func (l *loggerImpl) Alert(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Alert(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	ErrorLog.Alert(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}

func (l *loggerImpl) Error(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Error(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	ErrorLog.Error(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}

func (l *loggerImpl) Warn(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Warn(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	ErrorLog.Warn(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}

func (l *loggerImpl) Notice(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Notice(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	if strings.HasPrefix(format, "[pid:") {
		AccessLog.Notice(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
		return
	}
	AppLog.Notice(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}

func (l *loggerImpl) Info(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Info(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	AppLog.Info(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}

func (l *loggerImpl) Debug(format string, a ...interface{}) {
	if l.Ctx.IsDebug() {
		ConsoleLog.Debug(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
	}
	AppLog.Debug(l.Ctx.RequestID(), fmt.Sprintf(format, a...))
}
