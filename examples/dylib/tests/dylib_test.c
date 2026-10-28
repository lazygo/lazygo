// hostbridge 共享库 C 接口测试（仅依赖公共头 hostbridge.h，模拟真实宿主）。
//
// 编译（在 examples/dylib 目录，已构建 libhostbridge.dylib 于 /tmp）：
//   cc -I. tests/dylib_test.c -DUNICODE=1 -L/tmp -lhostbridge -o /tmp/dylib_test
//   DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test            # EVENT_LOOP 模式（推荐，生产同款）
//   DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test callback   # CALLBACK 模式
// 依赖：curl（HTTP 回环测试用，macOS/Linux 自带）。
// 窄字符库构建：-DUNICODE=0（必须与库的构建保持一致，见 hostbridge.h 头注释）。
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#ifndef UNICODE
#define UNICODE 1
#endif
#include "../hostbridge.h"

#if defined(UNICODE) && UNICODE
typedef wchar_t tchar_t;
    #define LIT(s) L##s
    #define FMT_S "%ls"
    #define tstr_cmp wcscmp
#else
typedef char tchar_t;
    #define LIT(s) s
    #define FMT_S "%s"
    #define tstr_cmp strcmp
#endif

#define TEST_PORT 18099
// Init 完成事件（main.go 中的 InitRID）
#define RID_INITIALIZED 2

static int g_pass = 0, g_fail = 0;
#define CHECK(cond, name) do { \
    if (cond) { g_pass++; printf("[PASS] %s\n", name); } \
    else      { g_fail++; printf("[FAIL] %s\n", name); } \
} while (0)

static volatile int g_got_initialized = 0;
static volatile int g_host_done = 0;

// CALLBACK 模式回调（签名随库构建：宽字符库收 wchar_t*，窄字符库收 char*）
#if defined(UNICODE) && UNICODE
static void test_callback(uint64_t rid, const wchar_t* uri, const wchar_t* body) {
#else
static void test_callback(uint64_t rid, const char* uri, const char* body) {
#endif
    printf("[Callback] rid: %llu, uri: " FMT_S ", body: " FMT_S "\n",
           (unsigned long long)rid, uri ? uri : LIT("(null)"), body ? body : LIT("(null)"));
    if (rid == RID_INITIALIZED)
        g_got_initialized = 1;
}

// 模拟宿主：轮询 EventLoop 取请求并处理。
// 生产流程：HTTP 客户端 -> 库内路由 -> 处理器经 Event.Request 投递到此处 -> 宿主用同一 rid AsyncSend 回送。
static void* host_emulator(void* arg) {
    (void)arg;
    tchar_t uri[256], resp[4096];
    for (int i = 0; i < 30000 && !g_host_done; i++) { // ~30s 安全上限
        memset(uri, 0, sizeof(uri));
        memset(resp, 0, sizeof(resp));
        uint64_t rid = EventLoop(uri, 256, resp, 4096);
        if (rid == 0) {
            usleep(1000);
            continue;
        }
        printf("[HostEmu] rid: %llu, uri: " FMT_S ", body: " FMT_S "\n",
               (unsigned long long)rid, uri, resp);
        if (tstr_cmp(uri, LIT("echo")) == 0) {
            AsyncSend(rid, LIT("echo"), LIT(""), LIT("\"pong\""));
        }
        // 其余事件（如库内推送）仅打印，不回送
    }
    return NULL;
}

// HTTP 回环：GET /api/rest/echo 经库路由到宿主，响应应包含宿主回送的 pong
static void test_http_roundtrip(void) {
    pthread_t tid;
    pthread_create(&tid, NULL, host_emulator, NULL);

    char cmd[256];
    snprintf(cmd, sizeof(cmd),
             "curl -s -m 15 'http://127.0.0.1:%d/api/rest/echo'", TEST_PORT);
    printf("[HTTP] %s\n", cmd);
    FILE* fp = popen(cmd, "r");
    char body[4096] = {0};
    if (fp) {
        fread(body, 1, sizeof(body) - 1, fp);
        pclose(fp);
    }
    printf("[HTTP] 响应: %s\n", body);
    CHECK(strstr(body, "pong") != NULL, "HTTP 请求经宿主回环拿到响应");

    g_host_done = 1;
    pthread_join(tid, NULL);
}

int main(int argc, char** argv) {
    int callback_mode = argc > 1 && strcmp(argv[1], "callback") == 0;
    printf("=== dylib_test (%s 字符模式) ===\n",
#if defined(UNICODE) && UNICODE
           "宽"
#else
           "窄"
#endif
    );

    int ret = Init(LIT("127.0.0.1"), TEST_PORT, NULL,
                   callback_mode ? DELIVERY_MODE_CALLBACK : DELIVERY_MODE_EVENT_LOOP,
                   callback_mode ? (uintptr_t)test_callback : 0);
    CHECK(ret == 0, "Init 返回 0");

    if (callback_mode) {
        CHECK(g_got_initialized, "Init 完成经回调投递 initialized 事件");
        printf("=== 通过 %d 项，失败 %d 项 ===\n", g_pass, g_fail);
        return g_fail > 0;
    }

    // Init 完成事件（rid=RID_INITIALIZED, uri="initialized"）应已入队，可被 EventLoop 取出
    tchar_t uri[256], resp[4096];
    memset(uri, 0, sizeof(uri));
    memset(resp, 0, sizeof(resp));
    uint64_t rid = EventLoop(uri, 256, resp, 4096);
    CHECK(rid == RID_INITIALIZED && tstr_cmp(uri, LIT("initialized")) == 0,
          "Init 完成事件经 EventLoop 取出");

    // 队列已空时应返回 0
    rid = EventLoop(uri, 256, resp, 4096);
    CHECK(rid == 0, "空队列 EventLoop 返回 0");

    // C-API Request 的语义：把事件派发给库内部 HTTP 路由（/event/<uri>）并同步返回其结果，
    // 不经过宿主——宿主只收到 REST/WS 处理器经 Event.Request 发来的请求（见 test_http_roundtrip）。
    tchar_t result[512] = {0};
    int req_ret = Request(42, LIT("test.api"), LIT(""), LIT("{\"msg\":\"ping\"}"), result, 511);
    printf("[Debug] C-API Request 响应: " FMT_S "\n", result);
    CHECK(req_ret == 0, "C-API Request 内部路由同步返回");

    // 生产全链路：HTTP 客户端 -> 库 -> EventLoop -> 宿主 AsyncSend -> HTTP 响应
    test_http_roundtrip();

    printf("=== 通过 %d 项，失败 %d 项 ===\n", g_pass, g_fail);
    return g_fail > 0;
}
