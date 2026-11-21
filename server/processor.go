package server

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/lazygo/pkg/goutils"
)

// processorEntry 预处理器，统一归一化为带可变参数的函数形态
type processorEntry struct {
	fn func(str string, args ...string) string
}

var (
	processorMu sync.RWMutex
	processors  = map[string]processorEntry{}
)

// RegisterProcessor 注册bind的process预处理器，供框架外部扩展
// name 不区分大小写，process注解中以 name 或 name(a,b) 形式引用
// fn 支持两种形态：
//   - func(string) string：不带额外参数，对应 process:"name"
//   - func(string, ...string) string：带额外参数，对应 process:"name(a,b)"，args 为括号内按逗号拆分的参数
//
// 同名重复注册时后者覆盖前者。注册非法签名会panic。
func RegisterProcessor(name string, fn any) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		panic("processor name is empty")
	}

	processorMu.Lock()
	defer processorMu.Unlock()
	switch f := fn.(type) {
	case func(string) string:
		processors[name] = processorEntry{fn: func(str string, _ ...string) string { return f(str) }}
	case func(string, ...string) string:
		processors[name] = processorEntry{fn: f}
	default:
		panic(fmt.Sprintf("unsupported processor signature, need func(string) string or func(string, ...string) string, got %T", fn))
	}
}

// parseProc 解析 process 指令，"name" 返回无参形式，"name(a,b)" 返回括号内按逗号拆分的参数
func parseProc(proc string) (string, []string) {
	i := strings.IndexByte(proc, '(')
	if i < 0 || !strings.HasSuffix(proc, ")") {
		return proc, nil
	}
	name := strings.TrimSpace(proc[:i])
	inner := proc[i+1 : len(proc)-1]
	if strings.TrimSpace(inner) == "" {
		return name, nil
	}
	parts := strings.Split(inner, ",")
	args := make([]string, 0, len(parts))
	for _, p := range parts {
		args = append(args, strings.TrimSpace(p))
	}
	return name, args
}

func process(str string, procList []string) string {
	for _, proc := range procList {
		proc = strings.TrimSpace(strings.ToLower(proc))
		if proc == "" {
			continue
		}
		name, args := parseProc(proc)
		processorMu.RLock()
		entry, ok := processors[name]
		processorMu.RUnlock()
		if !ok {
			continue
		}
		str = entry.fn(str, args...)
	}
	return str
}

func init() {
	RegisterProcessor("trim", func(str string) string {
		return strings.TrimSpace(str)
	})
	RegisterProcessor("tolower", func(str string) string {
		return strings.ToLower(str)
	})
	RegisterProcessor("toupper", func(str string) string {
		return strings.ToUpper(str)
	})
	RegisterProcessor("cut", func(str string, args ...string) string {
		if len(args) == 0 {
			return str
		}
		if n, err := strconv.Atoi(args[0]); err == nil {
			str = goutils.CutRune(str, n)
		}
		return str
	})
}
