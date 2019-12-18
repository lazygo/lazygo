package memcache

import (
	"fmt"
	"strings"
)

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
