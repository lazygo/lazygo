package utils

import (
	"fmt"
	"runtime/debug"
)

// CheckError 如果err != nil 记录日志
func CheckError(err error) {
	if err != nil {
		fmt.Println(string(debug.Stack()))
		fmt.Println(err)
	}
}
