package utils

import (
	"fmt"
	"runtime/debug"
)

// CheckError 如果err != nil 记录日志
func CheckError(err error) {
	if err != nil {
		fmt.Println(string(debug.Stack()))
		Warn(err.Error())
	}
}

// CheckFatal 如果err != nil 记录日志并panic
func CheckFatal(err error) {
	if err != nil {
		Error(err.Error())
		panic(err)
	}
}

// Go 安全的goroutine
func Go(callback func()) {
	go func() {
		defer func() { // 防止程序异常退出
			if e := recover(); e != nil {
				if err, ok := e.(error); ok {
					Warn(err.Error())
				} else {
					Warn("error", map[string]interface{}{"error": e})
				}
			}
		}()
		callback()
	}()
}
