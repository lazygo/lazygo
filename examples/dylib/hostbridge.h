#ifndef __HOSTBRIDGE_H__
#define __HOSTBRIDGE_H__

// 本库（Go/lazygo 实现，编译为共享库）对宿主的唯一公共头文件，自包含，不依赖任何内部头文件。
//
// 字符宽度：库的 ABI 随构建时的 UNICODE 开关为宽字符（wchar_t）或窄字符（char，UTF-8）。
// 包含本头文件前必须令 UNICODE 与库的构建保持一致（宽字符库先 #define UNICODE 1）——
// C 链接按符号名解析、不做类型检查，原型与库不符时不会报错但数据会错。

#include <stdint.h>
#include <wchar.h>

#ifdef __cplusplus
extern "C" {
#endif

// 请求投递模式（Init 的 mode 参数取值）：选定「本库如何把收到的 HTTP/WS 请求交给宿主执行」
// 的线程安全投递机制。以宏常量提供、不暴露枚举类型——C ABI 上与 int 完全等价，直接传数值亦可。
//
// 背景：库内置的 HTTP/WS 服务器跑在 Go 自己的 goroutine 上（对宿主是外来线程），而宿主
// 代码通常有线程亲和性限制，不能从外来线程回调，故每个请求必须经以下三种机制之一投递。
// 响应回程与投递方式无关：宿主处理完统一用 AsyncSend(rid, ...) 回送，靠 rid 对上挂起的请求。
#define DELIVERY_MODE_CALLBACK 0      // 直接回调：handle 为 C 函数指针，库直接调用；回调在 Go goroutine 上执行，仅适用于可从任意线程重入的纯 C 宿主
#define DELIVERY_MODE_SEND_MESSAGE 1  // WM_COPYDATA：handle 为宿主窗口句柄，SendMessage 同步投递；仅 Windows、需宿主有窗口和消息循环
#define DELIVERY_MODE_EVENT_LOOP 2    // 事件队列轮询：宿主主循环反复调用 EventLoop() 取件；无跨线程调用，脚本/服务宿主唯一可行（推荐）

// 形参 const 仅约束调用方的类型检查，不影响 ABI。本头文件独立于库编译（不进入 cgo
// 生成导出台词的编译单元），故输入指针保留 const；输出缓冲区（resp 等）不带 const。
#if defined(UNICODE) && UNICODE

extern int Init(const wchar_t* listenAddr, int listenPort, const wchar_t* dataPath, int mode, uintptr_t handle);
extern int AsyncSend(uint64_t rid, const wchar_t* uri, const wchar_t* header, const wchar_t* body);
extern int Request(uint64_t rid, const wchar_t* uri, const wchar_t* header, const wchar_t* body, wchar_t* resp, unsigned int szResp);
extern uint64_t EventLoop(wchar_t* uri, unsigned int szURI, wchar_t* resp, unsigned int szResp);

#else

extern int Init(const char* listenAddr, int listenPort, const char* dataPath, int mode, uintptr_t handle);
extern int AsyncSend(uint64_t rid, const char* uri, const char* header, const char* body);
extern int Request(uint64_t rid, const char* uri, const char* header, const char* body, char* resp, unsigned int szResp);
extern uint64_t EventLoop(char* uri, unsigned int szURI, char* resp, unsigned int szResp);

#endif

#ifdef __cplusplus
}
#endif

#endif /* __HOSTBRIDGE_H__ */
