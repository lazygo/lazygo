# dylib 示例：把 lazygo 服务编译为 C 动态库

本示例演示如何用 lazygo 搭建的 HTTP 服务编译成 **C 共享库**（`-buildmode=c-shared`），
被 C 程序、脚本引擎（如 MQL）等**宿主**加载：库内置 HTTP/WS 服务器跑在 Go 自己的
goroutine 上，宿主通过四个 C API 与库双向交互，请求经三种**投递模式**之一线程安全地
交给宿主执行（详见 `main.h` 中 `DeliveryMode` 的设计注释）。

## 宿主 API（hostbridge.h）

| 函数 | 方向 | 说明 |
|---|---|---|
| `Init(addr, port, dataDir, mode, handle)` | 宿主→库 | 启动服务；`mode` 取 `DELIVERY_MODE_*`，`handle` 随模式为回调函数指针或窗口句柄 |
| `EventLoop(uri, szURI, resp, szResp)` | 库→宿主 | 事件队列模式：宿主主循环轮询取请求（返回 rid，0 表示空） |
| `AsyncSend(rid, uri, header, body)` | 宿主→库 | 宿主回送/推送事件，按 rid 配对挂起的请求 |
| `Request(rid, uri, header, body, resp, szResp)` | 宿主→库 | 同步调用：派发给库内部 `/event/<uri>` 路由并返回其结果（不经过宿主） |

## 结构

```text
.
├── main.go          // //export 宿主 API 入口（Init/AsyncSend/Request/EventLoop）
├── bridge.go        // callHandle 投递分发 + TCHAR↔Go string 转换
├── server.go        // lazygo 服务启动 + echo 演示路由 + 事件桥（CallWrapper）
├── main.h/main.c    // 内部 C 头与助手（TCHAR 宏、InvokeCallback/InvokeSend、CopyString）
├── hostbridge.h     // 对外公共头（自包含，宿主侧只需它）
└── tests/dylib_test.c  // C 接口测试（EventLoop/CALLBACK 双模式、HTTP 全链路回环）
```

## 构建

依赖：Go 1.27+、CGO；交叉编译 Windows 需 mingw-w64（如 llvm-mingw）。

```bash
# macOS
CGO_ENABLED=1 go build -buildmode=c-shared -ldflags "-w -s" -trimpath -o libhostbridge.dylib .

# Windows（交叉编译示例，CC 指向 mingw clang）
CGO_ENABLED=1 CC=x86_64-w64-mingw32-clang GOOS=windows GOARCH=amd64 \
  go build -buildmode=c-shared -ldflags "-w -s -extldflags '-static'" -trimpath -o hostbridge.dll .

# Linux
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  go build -buildmode=c-shared -ldflags "-w -s" -trimpath -o libhostbridge.so .
```

构建同时生成同名 `.h`（cgo 导出声明）。对外契约以 `hostbridge.h` 为准。

## 字符宽度（UNICODE 开关）

`main.h` 的 `#define UNICODE 1` 决定 ABI：宽字符（`wchar_t`，MQL 等宿主必须）
或窄字符（`char`，UTF-8，纯 C 宿主）。宿主包含 `hostbridge.h` 时须用 `-DUNICODE=`
保持与库的构建一致。

## 运行测试

```bash
# 构建 dylib 到 /tmp 后：
cc -I. tests/dylib_test.c -DUNICODE=1 -L/tmp -lhostbridge -o /tmp/dylib_test
DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test            # EVENT_LOOP 模式（生产同款）
DYLD_LIBRARY_PATH=/tmp /tmp/dylib_test callback   # CALLBACK 模式（函数指针回调）
```
