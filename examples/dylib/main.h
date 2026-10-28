#ifndef __MAIN_H__
#define __MAIN_H__

#include <stdlib.h>
#include <stdint.h>
#include <wchar.h>


// 字符宽度开关：UNICODE=1 全库使用宽字符（wchar_t），UNICODE=0 或不定义则使用窄字符（char，按 UTF-8 解释）。
// 注意：宿主字符串若是宽字符（如 MQL 的 UTF-16 string），构建对应动态库时必须保持 UNICODE=1；
// 窄字符构建面向 char/UTF-8 宿主（如纯 C 程序）。
#define UNICODE 1

// 请求投递模式：选定「本库（Go 实现，编译为共享库）如何把收到的 HTTP/WS 请求交给
// 宿主程序执行」的线程安全投递机制。
//
// 背景：本库以共享库（DLL/dylib/so）形式被宿主程序加载，内置 HTTP/WS 服务器跑在
// Go 自己的 goroutine 上（对宿主而言是外来线程）；而宿主代码通常有线程亲和性限制——
// 只能在宿主自身线程执行，不允许从外来线程回调（GUI 消息循环、专有 API 运行时等
// 均如此）。因此每个请求必须经以下三种机制之一投递给宿主，Init 时传入，由库内
// callHandle 按此 switch 分发。
// 响应回程与投递方式无关：宿主处理完统一用 AsyncSend(rid, ...) 回送，靠 rid 对上
// 库内挂起的请求。
//
// 三种模式分别适配不同形态的宿主，故全部保留：宿主是可从任意线程重入的纯 C 代码
// 时用 CALLBACK；宿主是有窗口/消息循环的 GUI 程序时用 SEND_MESSAGE；宿主是脚本/服务、
// 只能自持主循环轮询时用 EVENT_LOOP。
typedef enum {
  // 直接回调：handle 为 C 函数指针，库将其强转为 Callback 直接调用。
  // 回调在 Go goroutine 上执行，仅适用于可从任意线程安全重入的纯 C 宿主
  // （如 C 测试程序、经 FFI 传入回调的其他语言集成）；
  // 有线程亲和性限制的宿主不可用，无法把函数指针作为整数句柄传入的宿主也不可用。
  DELIVERY_MODE_CALLBACK = 0,

  // WM_COPYDATA：handle 为宿主窗口句柄（宿主用其窗口 API 获取，如 Win32 的
  // WindowHandle()），库用 SendMessage(hwnd, WM_COPYDATA, rid, &cds) 投递。
  // Windows 消息机制线程安全；但 SendMessage 同步阻塞库的 handler 直至宿主处理完毕，
  // 且仅 Windows 可用、要求宿主有窗口和消息循环（无窗口的脚本/服务宿主不适用）。
  DELIVERY_MODE_SEND_MESSAGE = 1,

  // 事件队列轮询：库将 {rid, uri, body} 入队，宿主在自己的主循环中反复调用
  // EventLoop() 取出并分派。
  // 完全无跨线程调用、跨平台，是无窗口、无消息循环的宿主（脚本/服务形态）唯一
  // 可行的模式；代价是轮询间隔级的投递延迟，且宿主必须自持主循环。
  DELIVERY_MODE_EVENT_LOOP = 2,
} DeliveryMode;

// 让 CopyString/tcslen 映射为宽字符或非宽字符版本
#if defined(UNICODE) && UNICODE
    #define CopyString  CopyStringW
    #define TCHAR wchar_t
    #define tcslen wcslen
#else
    #define CopyString  CopyStringA
    #define TCHAR char
    #define tcslen strlen
#endif

// 让 Callback 映射为宽字符或非宽字符函数
typedef void (*Callback)(uint64_t rid, const TCHAR *uri, const TCHAR *body);


/* C 端函数声明，供 Go/cgo 调用 */
void InvokeCallback(uintptr_t handle, uint64_t rid, const TCHAR* uri, const TCHAR* body);
void CopyStringA(char* dest, const char* src, unsigned int size);
void CopyStringW(wchar_t* dest, const wchar_t* src, unsigned int size);
int InvokeSend(uintptr_t handle, uint64_t rid, const TCHAR* uri, const TCHAR* body);

#endif /* __MAIN_H__ */
