package server

import (
	stdContext "context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strconv"

	"github.com/coder/websocket"
	"github.com/lazygo/pkg/waiter"
)

var (
	_ SendReceiveCloser = (*wsBridge)(nil)
	_ SendReceiveCloser = (*callBridge)(nil)
)

// WrapHandler wraps `http.Handler` into `HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx.SetRequest(r)
				ctx.SetResponseWriter(NewResponseWriter(w))
				err = next(ctx)
			})).ServeHTTP(ctx.ResponseWriter(), ctx.Request())
			return
		}
	}
}

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

// AssetHandler returns an http.Handler that will serve files from
// the Assets embed.FS. When locating a file, it will strip the given
// prefix from the request and prepend the root to the filesystem.
func AssetHandler(prefix string, assets embed.FS, root string) HandlerFunc {
	handler := fsFunc(func(name string) (fs.File, error) {
		assetPath := path.Join(root, name)

		// If we can't find the asset, fs can handle the error
		file, err := assets.Open(assetPath)
		if err != nil {
			return nil, err
		}

		// Otherwise assume this is a legitimate request routed correctly
		return file, err
	})

	return WrapHandler(http.StripPrefix(prefix, http.FileServer(http.FS(handler))))
}

func WebSocketWrapper(ctx stdContext.Context, conn *websocket.Conn) SendReceiveCloser {
	return &wsBridge{ctx: ctx, conn: conn}
}

type wsBridge struct {
	ctx  stdContext.Context
	conn *websocket.Conn
}

func (b *wsBridge) Receive(ctx stdContext.Context) (*EventData, error) {

	_, msg, err := b.conn.Read(b.ctx)
	if err != nil {
		return nil, err
	}

	var req EventData
	err = json.Unmarshal(msg, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (b *wsBridge) Send(data *EventData) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = b.conn.Write(b.ctx, websocket.MessageText, msg)
	if err != nil {
		return err
	}
	return nil
}

func (b *wsBridge) Close() error {
	return b.conn.Close(websocket.StatusNormalClosure, "normal closure")
}

func CallWrapper(ctx stdContext.Context, callback func(*EventData)) CallBridge {
	return &callBridge{
		ctx:      ctx,
		ch:       make(chan *EventData),
		callback: callback,
		waiter:   waiter.NewWaiter[*EventData](),
	}
}

type CallBridge interface {
	SendReceiveCloser
	SendRequest(ctx stdContext.Context, req *EventData) error
	Waiter(ctx stdContext.Context, id string) func() (*EventData, error)
}

type callBridge struct {
	ctx      stdContext.Context
	ch       chan *EventData
	callback func(*EventData)
	waiter   *waiter.Waiter[*EventData]
}

func (b *callBridge) SendRequest(ctx stdContext.Context, req *EventData) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case b.ch <- req:
		return nil
	}
}

func (b *callBridge) Waiter(ctx stdContext.Context, id string) func() (*EventData, error) {
	return b.waiter.Get(ctx, id)
}

func (b *callBridge) Receive(ctx stdContext.Context) (*EventData, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data, ok := <-b.ch:
		if !ok {
			return nil, errors.New("channel closed")
		}
		return data, nil
	}
}

func (b *callBridge) Send(req *EventData) error {
	// 解析数据判断是给cb还是写入给pipe
	if req.RID > 0 {
		ok, err := b.waiter.Put(b.ctx, strconv.FormatUint(req.RID, 10), req)
		if err != nil {
			return err
		}
		if ok {
			// 写入成功，直接返回
			return nil
		}
	}
	// rid = 0 或 put失败，则写入pipe
	b.callback(req)
	return nil
}

func (b *callBridge) Close() error {
	close(b.ch)
	return nil
}
