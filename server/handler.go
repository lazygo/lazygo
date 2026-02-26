package server

import (
	stdContext "context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strconv"

	"github.com/coder/websocket"
	"github.com/lazygo/pkg/receiver"
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

func WebSocketWrapper(ctx stdContext.Context, conn *websocket.Conn) io.ReadWriteCloser {
	return &wsBridge{ctx: ctx, conn: conn}
}

type wsBridge struct {
	ctx  stdContext.Context
	conn *websocket.Conn
	// buffer stores remaining unread data from the last msg
	buffer []byte
}

func (b *wsBridge) Read(p []byte) (n int, err error) {
	// If buffer has leftover data, serve from buffer first
	if len(b.buffer) > 0 {
		n = copy(p, b.buffer)
		b.buffer = b.buffer[n:]
		return n, nil
	}

	io.Pipe()
	net.Pipe()
	_, msg, err := b.conn.Read(b.ctx)
	if err != nil {
		return 0, err
	}

	n = copy(p, msg)
	// If msg not fully read, save the rest into buffer
	if n < len(msg) {
		b.buffer = append(b.buffer[:0], msg[n:]...)
	} else {
		b.buffer = b.buffer[:0]
	}
	return n, nil
}

func (b *wsBridge) Write(p []byte) (n int, err error) {
	// 这里已作出相应处理保证p为一次完整的响应，不会被截断
	err = b.conn.Write(b.ctx, websocket.MessageText, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (b *wsBridge) Close() error {
	return b.conn.Close(websocket.StatusNormalClosure, "normal closure")
}

func CallWrapper(ctx stdContext.Context, callback func(rid uint64, uri string, body []byte)) *callBridge {
	pipeReader, pipeWriter := io.Pipe()
	return &callBridge{
		ctx:        ctx,
		pipeReader: pipeReader,
		pipeWriter: pipeWriter,
		callback:   callback,
		receiver:   receiver.NewReceiver[[]byte](),
	}
}

type callBridge struct {
	ctx        stdContext.Context
	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
	callback   func(rid uint64, uri string, body []byte)
	receiver   *receiver.Receiver[[]byte]
}

func (b *callBridge) PipeWriter() io.Writer {
	return b.pipeWriter
}

func (b *callBridge) Receiver(ctx stdContext.Context, id string) func() ([]byte, error) {
	return b.receiver.Get(ctx, id)
}

func (b *callBridge) Read(p []byte) (n int, err error) {
	return b.pipeReader.Read(p)
}

func (b *callBridge) Write(p []byte) (n int, err error) {
	// 解析数据判断是给cb还是写入给pipe
	req := EventRequest{}
	err = json.Unmarshal(p, &req)
	if err != nil {
		return 0, err
	}
	if req.RID > 0 {
		ok, err := b.receiver.Put(b.ctx, strconv.FormatUint(req.RID, 10), p)
		if err != nil {
			return 0, err
		}
		if ok {
			// 写入成功，直接返回
			return len(p), nil
		}
	}
	// rid = 0 或 put失败，则写入pipe
	b.callback(req.RID, req.URI, req.Body)
	return len(req.Body), nil
}

func (b *callBridge) Close() error {
	b.pipeWriter.Close()
	b.pipeReader.Close()
	return nil
}
