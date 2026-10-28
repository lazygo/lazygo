package main

/*
#cgo CFLAGS: -Wall -Wextra -Werror -fPIC
#cgo CXXFLAGS: -Wall -Wextra -Werror -fPIC
#include "main.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"net/http"
	"unicode/utf16"
	"unsafe"

	"github.com/lazygo/lazygo/server"
)

func callHandle(handle C.uintptr_t, rid uint64, uri string, reqBody []byte) error {
	switch DeliveryMode {
	case C.DELIVERY_MODE_CALLBACK:
		cURI := GoStringTChar(uri)
		defer C.free(unsafe.Pointer(cURI))
		cBody := GoStringTChar(string(reqBody))
		defer C.free(unsafe.Pointer(cBody))
		C.InvokeCallback(handle, C.uint64_t(rid), cURI, cBody)
		return nil
	case C.DELIVERY_MODE_SEND_MESSAGE:
		cURI := GoStringTChar(uri)
		defer C.free(unsafe.Pointer(cURI))
		cBody := GoStringTChar(string(reqBody))
		defer C.free(unsafe.Pointer(cBody))
		ret := C.InvokeSend(handle, C.uint64_t(rid), cURI, cBody)
		if ret != 0 {
			return fmt.Errorf("send message fail: %d", ret)
		}
		return nil
	case C.DELIVERY_MODE_EVENT_LOOP:
		event := &server.EventData{
			RID:    rid,
			URI:    uri,
			Header: http.Header{},
			Body:   reqBody,
		}
		select {
		default:
			// 清理eventQueue，防止eventQueue满
			select {
			case <-eventQueue:
			default:
			}
			return errors.New("event queue is full")
		case eventQueue <- event:
			return nil
		}
	}
	return errors.New("invalid mode")
}

// tcharWidth 是 C 侧 TCHAR 的宽度（字节数）：1=char（按 UTF-8 解释），2=UTF-16，4=UTF-32。
// 由库编译时的 UNICODE 开关决定，Go 侧借此在窄/宽字符两种 ABI 间复用同一套代码。
var tcharWidth = unsafe.Sizeof(*new(C.TCHAR))

// TCharGoString 将 C 的 TCHAR* 转换为 Go string
func TCharGoString(p *C.TCHAR) string {
	if p == nil {
		return ""
	}
	switch tcharWidth {
	case 1:
		return C.GoString((*C.char)(unsafe.Pointer(p)))
	case 2, 4:
		return wcharGoString((*C.wchar_t)(unsafe.Pointer(p)))
	}
	return ""
}

// GoStringTChar 将 Go string 转换为 C 的 TCHAR*，窄字符按 UTF-8 编码
// 返回的指针需要调用方在使用后使用 C.free 释放
func GoStringTChar(s string) *C.TCHAR {
	if tcharWidth == 1 {
		if s == "" {
			// 返回只包含 '\0' 的空串指针
			return (*C.TCHAR)(C.calloc(1, 1))
		}
		return (*C.TCHAR)(unsafe.Pointer(C.CString(s)))
	}
	return (*C.TCHAR)(unsafe.Pointer(goStringWchar(s)))
}

func wcharGoString(wcs *C.wchar_t) string {
	if wcs == nil {
		return ""
	}
	length := C.wcslen(wcs)
	if length == 0 {
		return ""
	}

	// 判断 wchar_t 大小
	sz := unsafe.Sizeof(*wcs)
	switch sz {
	case 2: // UTF-16
		buf := make([]uint16, length)
		for i := C.int(0); i < C.int(length); i++ {
			buf[i] = uint16(*(*C.wchar_t)(unsafe.Add(unsafe.Pointer(wcs), uintptr(i)*sz)))
		}
		runes := utf16.Decode(buf)
		return string(runes)
	case 4: // UTF-32
		buf := make([]uint32, length)
		for i := C.int(0); i < C.int(length); i++ {
			buf[i] = uint32(*(*C.wchar_t)(unsafe.Add(unsafe.Pointer(wcs), uintptr(i)*sz)))
		}
		// 直接转换
		runes := make([]rune, length)
		for i, v := range buf {
			runes[i] = rune(v)
		}
		return string(runes)
	default:
		// 其他情况（极少见）可以按字节复制并尝试用 UTF-8 解码，或返回错误
		return ""
	}
}

// goStringWchar 将 Go string 转换为 C 的 wchar_t* (宽字符串)
// 返回的指针需要调用方在使用后使用 C.free 释放
func goStringWchar(s string) *C.wchar_t {
	if s == "" {
		// 返回空字符串的指针，即只包含 L'\0'
		return (*C.wchar_t)(C.calloc(1, C.sizeof_wchar_t))
	}

	// 将 Go string 转为 rune 切片
	runes := []rune(s)

	// 判断 wchar_t 大小
	sz := unsafe.Sizeof(*new(C.wchar_t))
	switch sz {
	case 2: // UTF-16
		// 将 rune 切片编码为 UTF-16 码元序列 ([]uint16)
		utf16Codes := utf16.Encode(runes)
		// elemCount = C.int(len(utf16Codes))

		// 计算所需字节数并分配 C 内存
		byteCount := (len(utf16Codes) + 1) * int(sz) // +1 用于 L'\0'
		cBuf := C.malloc(C.size_t(byteCount))
		if cBuf == nil {
			panic("C.malloc failed")
		}

		// 将 uint16 切片转为字节序列（按主机字节序）并拷贝
		// 注意：Go 的 uint16 在内存中就是小端表示，但 C 的 wchar_t 可能要求主机字节序，所以直接拷贝即可
		for i, code := range utf16Codes {
			// 计算目标位置
			dest := unsafe.Pointer(uintptr(cBuf) + uintptr(i)*sz)
			// 将 uint16 写入两个字节（根据主机字节序）
			*(*C.wchar_t)(dest) = C.wchar_t(code)
		}
		// 添加终止符
		dest := unsafe.Pointer(uintptr(cBuf) + uintptr(len(utf16Codes))*sz)
		*(*C.wchar_t)(dest) = 0

		return (*C.wchar_t)(cBuf)

	case 4: // UTF-32
		// 直接将 rune 转换为 uint32 序列
		// elemCount = C.int(len(runes))
		byteCount := (len(runes) + 1) * int(sz)
		cBuf := C.malloc(C.size_t(byteCount))
		if cBuf == nil {
			panic("C.malloc failed")
		}

		for i, r := range runes {
			dest := unsafe.Pointer(uintptr(cBuf) + uintptr(i)*sz)
			*(*C.wchar_t)(dest) = C.wchar_t(r)
		}
		// 添加终止符
		dest := unsafe.Pointer(uintptr(cBuf) + uintptr(len(runes))*sz)
		*(*C.wchar_t)(dest) = 0

		return (*C.wchar_t)(cBuf)
	default:
		// 其他宽度不支持
		panic("unsupported wchar_t size")
	}
}
