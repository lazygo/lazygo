package main

// 本示例把 lazygo HTTP 服务编译为 C 动态库（c-shared），供 C/脚本等宿主加载调用。
// debug: go build -buildmode=c-shared -o "libhostbridge.dylib" .
// macos: CGO_ENABLED=1 CC=clang GOOS=darwin GOARCH=arm64 go build -buildmode=c-shared -buildvcs=false -ldflags "-w -s" -trimpath -o libhostbridge.dylib .
// windows: CGO_ENABLED=1 CC=x86_64-w64-mingw32-clang GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -buildvcs=false -ldflags "-w -s -extldflags '-static'" -trimpath -o hostbridge.dll .
// linux: CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildmode=c-shared -buildvcs=false -ldflags "-w -s" -trimpath -o libhostbridge.so .

/*
#cgo CFLAGS: -Wall -Wextra -Werror -fPIC
#cgo CXXFLAGS: -Wall -Wextra -Werror -fPIC
#include "main.h"
*/
import "C"

import (
	"cmp"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Llongfile | log.LstdFlags)
}

func main() {}

const (
	ErrorRID = 0x1
	InitRID  = 0x2
)

var DeliveryMode C.DeliveryMode

//export Init
func Init(listenAddr *C.TCHAR, listenPort C.int, dataPath *C.TCHAR, mode C.DeliveryMode, handle C.uintptr_t) C.int {
	DeliveryMode = mode
	serverAddr := fmt.Sprintf("%s:%d", cmp.Or(TCharGoString(listenAddr), "127.0.0.1"), cmp.Or(int(listenPort), 8080))
	dataDir := cmp.Or(TCharGoString(dataPath), filepath.Join(os.TempDir(), "lazygo-dylib-data"))

	cb := func(rid uint64, uri string, body []byte) {
		callHandle(handle, rid, uri, body)
	}

	err := initServer(serverAddr, dataDir, cb)
	if err != nil {
		callHandle(handle, ErrorRID, "error", []byte(err.Error()))
		return -1
	}

	fmt.Println("initServer ok")

	callHandle(handle, InitRID, "initialized", []byte("ok"))
	return 0
}

//export AsyncSend
func AsyncSend(rid C.uint64_t, uri *C.TCHAR, header *C.TCHAR, body *C.TCHAR) C.int {
	// Copy C pointers to Go memory *before* starting goroutine:
	// after return the caller (host) may reuse/free the buffers, so using them in goroutine causes sigpanic.
	goRID := uint64(rid)
	goURI := TCharGoString(uri)
	reqBody := []byte(TCharGoString(body))
	headerBytes := []byte(TCharGoString(header))
	var headers http.Header
	if len(headerBytes) > 0 {
		hrader, err := url.ParseQuery(string(headerBytes))
		if err != nil {
			log.Printf("[msg: parse header error] [err: %v]", err)
		}
		headers = http.Header(hrader)
	}
	go func() {
		// 宿主 -> 库方向：把宿主产生的事件/请求推入库内事件系统（等待同 rid 的回送，
		// 若无人接收则超时丢弃）。如需响应，宿主侧应使用 Request。
		_, _ = request(goRID, goURI, headers, reqBody)
	}()
	return 0
}

//export Request
func Request(rid C.uint64_t, uri *C.TCHAR, header *C.TCHAR, body *C.TCHAR, resp *C.TCHAR, szResp C.uint) C.int {
	headerBytes := []byte(TCharGoString(header))
	var headers http.Header
	if len(headerBytes) > 0 {
		hrader, err := url.ParseQuery(string(headerBytes))
		if err != nil {
			return -1
		}
		headers = http.Header(hrader)
	}
	reqBody := []byte(TCharGoString(body))
	respBody, err := request(uint64(rid), TCharGoString(uri), headers, reqBody)
	if err != nil {
		cErr := GoStringTChar(err.Error())
		defer C.free(unsafe.Pointer(cErr))
		C.CopyString(resp, cErr, szResp)
		return -2
	}
	if len(respBody) == 0 {
		cErr := GoStringTChar("request fail")
		defer C.free(unsafe.Pointer(cErr))
		C.CopyString(resp, cErr, szResp)
		return -3
	}

	cBody := GoStringTChar(string(respBody))
	defer C.free(unsafe.Pointer(cBody))
	C.CopyString(resp, cBody, szResp)
	return 0
}

//export EventLoop
func EventLoop(uri *C.TCHAR, szURI C.uint, resp *C.TCHAR, szResp C.uint) C.uint64_t {
	select {
	default:
		return 0
	case e := <-eventQueue:
		if e == nil {
			return 0
		}
		cURI := GoStringTChar(e.URI)
		defer C.free(unsafe.Pointer(cURI))
		cBody := GoStringTChar(string(e.Body))
		defer C.free(unsafe.Pointer(cBody))
		C.CopyString(uri, cURI, szURI)
		C.CopyString(resp, cBody, szResp)
		return C.uint64_t(e.RID)
	}
}
