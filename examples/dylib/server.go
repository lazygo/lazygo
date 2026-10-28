package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/lazygo/lazygo/server"
)

// callSubject 是库内事件总线上「宿主调用」主题的标识
const callSubject = "lazygo-dylib-example"

var callBridge server.CallBridge
var eventQueue = make(chan *server.EventData, 2<<10)
var ridCounter atomic.Uint64

// initServer 启动 lazygo HTTP 服务，并挂接「HTTP <-> 宿主」双向事件桥
func initServer(listenAddr, dataDir string, cb func(rid uint64, uri string, respBody []byte)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic[initServer]: %v", r)
			log.Printf("[msg: panic] [err: %v]", err)
		}
	}()

	s := server.New()
	s.Group("/api/rest").Get("echo", echoHandler)

	ctx := context.Background()

	go func() {
		fmt.Println("Listen " + listenAddr)
		if err := s.Start(ctx, listenAddr); err != nil && err != http.ErrServerClosed {
			log.Printf("[msg: server stopped] [err: %v]", err)
		}
	}()

	// HTTP -> 宿主方向：REST/WS 处理器经 Event.Request 把请求投递给宿主（callHandle 按投递模式分发），
	// 宿主处理完用同一 rid 经 AsyncSend 回送，由这里的 waiter 配对返回给 HTTP 客户端
	callBridge = server.CallWrapper(ctx, func(e *server.EventData) {
		cb(e.RID, e.URI, e.Body)
	})
	event := s.Event(server.MethodCall, callSubject)
	go func() {
		if serveErr := event.Serve(ctx, 0, callBridge); serveErr != nil {
			log.Printf("[msg: event serve stopped] [err: %v]", serveErr)
		}
	}()

	_ = dataDir // 示例暂无持久化需求，保留参数以与通用 Init 签名一致
	return nil
}

// echoHandler 演示「HTTP -> 宿主 -> HTTP」全链路：
// 请求经事件总线投递给宿主，等待宿主回送后返回其响应体。
func echoHandler(c server.Context) error {
	rid := ridCounter.Add(1)
	data := &server.EventData{
		RID:    rid,
		URI:    "echo",
		Header: http.Header{},
		Body:   []byte("ping"),
	}
	ctx, cancel := context.WithTimeout(c, time.Minute)
	defer cancel()
	wait, err := c.Event(server.MethodCall, callSubject).Request(ctx, 0, data)
	if err != nil {
		return err
	}
	response, err := wait()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"rid":  rid,
		"body": string(response.Body),
	})
}

// request 宿主 -> 库方向：把宿主的事件/请求注入库内事件系统，并按 rid 等待响应。
// C-API Request（同步）与 AsyncSend（异步 goroutine）共用此实现。
func request(rid uint64, uri string, headers http.Header, body []byte) ([]byte, error) {
	req := &server.EventData{
		RID:    rid,
		URI:    path.Join("/event", uri),
		Header: headers,
		Body:   body,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	wait, cancel := callBridge.Waiter(ctx, strconv.FormatUint(rid, 10))
	err := callBridge.SendRequest(ctx, req)
	if err != nil {
		cancel()
		return nil, err
	}
	response, err := wait()
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}
