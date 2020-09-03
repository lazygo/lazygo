package utils

import (
	"fmt"
	"strings"
)

func CheckFatal(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func ContainInArray(str string, arr []string) bool {
	for _, s := range arr {
		if strings.Index(s, str) != -1 {
			return true
		}
	}
	return false
}

func InArray(str string, arr []string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
