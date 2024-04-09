package libhttp

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func toString(value any) string {
	var str string
	switch value.(type) {
	case string:
		str = value.(string)
	case []byte:
		str = string(value.([]byte))
	case int:
		str = strconv.Itoa(value.(int))
	case int64:
		str = strconv.FormatInt(value.(int64), 10)
	default:
		str = ""
	}
	return str
}

func recoverGo(callback func()) {
	go func() {
		defer func() { // 防止程序异常退出
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		callback()
	}()
}

// timeoutDialer returns functions of connection dialer with timeout settings for http.Transport Dial field.
func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		err = conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, err
	}
}
