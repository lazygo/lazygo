#ifdef _WIN32
#include <windows.h>
#endif
#include <string.h>

#include "main.h"

/* PE/COFF (MinGW) 不支持 ELF 式 visibility，GCC 会报 warning；Linux/macOS 上 GCC/Clang 均可用 */
#if (defined(__GNUC__) || defined(__clang__)) && !defined(_WIN32)
#define HIDDEN __attribute__((visibility("hidden")))
#else
#define HIDDEN
#endif

void HIDDEN InvokeCallback(uintptr_t handle, uint64_t rid, const TCHAR* uri, const TCHAR* body) {
    ((Callback)handle)(rid, uri, body);
}

// 宽字符（UTF-16/32，wchar_t）版本
void HIDDEN CopyStringW(wchar_t *dest, const wchar_t *src, unsigned int size) {
    // size 是 wchar_t 元素个数
    // 拷贝最多 size-1 个字符，末尾补 L'\0'
    if (size == 0) return;
#ifdef _WIN32
    wcsncpy_s(dest, size, src, _TRUNCATE);
#else
    size_t src_len = wcslen(src);
    size_t copy_len = src_len < (size - 1) ? src_len : (size - 1);
    wcsncpy(dest, src, copy_len);
    dest[copy_len] = L'\0';
#endif
}

// 非宽字符（普通 char*）版本
void HIDDEN CopyStringA(char *dest, const char *src, unsigned int size) {
    // size 是包含末尾 \0 的字节数
#ifdef _WIN32
    strncpy_s(dest, size, src, _TRUNCATE);
#else
    size_t src_len = strlen(src);
    size_t copy_len = src_len < (size - 1) ? src_len : (size - 1);
    strncpy(dest, src, copy_len);
    dest[copy_len] = '\0';
#endif
}

/**
 * 通过 WM_COPYDATA 向宿主窗口发送 TCHAR 字符串（窄字符为 UTF-8，宽字符为 UTF-16）。
 * 宿主在其窗口消息循环中接收 WM_COPYDATA 并分派（见 main.h 的 SEND_MESSAGE 模式）。
 *
 * @param handle    目标窗口句柄（由宿主用其窗口 API 获取并传入，如 Win32 的 WindowHandle()）
 * @param rid       请求 ID，将作为 wParam 传递（64 位整数）
 * @param uri       保留参数，未使用（可传 NULL）
 * @param body      要发送的 TCHAR 字符串（以 '\0'/L'\0' 结尾）
 * @return 0 成功，!= 0 失败
 */
int HIDDEN InvokeSend(uintptr_t handle, uint64_t rid, const TCHAR* uri, const TCHAR* body)
{
    (void)uri;  // 避免未使用参数的警告

#ifdef _WIN32

    // 准备 COPYDATASTRUCT
    COPYDATASTRUCT cds;
    cds.dwData = 1;                                       // 自定义数据类型标识
    cds.cbData = (DWORD)(tcslen(body) * sizeof(TCHAR));   // 数据总字节数（含终止符）
    cds.lpData = (void*)body;                             // 指向字符串数据的指针

    // 发送消息（同步）
    if (SendMessage((HWND)handle, WM_COPYDATA, (WPARAM)rid, (LPARAM)&cds) == 0) {
        return -1;
    }

#else
    // 非 Windows 平台：打印调试信息
#if defined(UNICODE) && UNICODE
    printf("SendMessage: %p %llu %ls %ls\n", (void*)handle, rid, uri ? uri : L"", body ? body : L"");
#else
    printf("SendMessage: %p %llu %s %s\n", (void*)handle, rid, uri ? uri : "", body ? body : "");
#endif
#endif
    return 0;
}
