package server

import (
	"bufio"
	"bytes"
	stdContext "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/coder/websocket"
)

var wsManager = &WSManager{}

type WSManager struct {
	ws sync.Map
}

func (w *WSManager) Get(name string) WebSocket {
	ws, _ := w.ws.LoadOrStore(name, &websocketImpl{
		ws: make(map[uint64]*websocket.Conn),
	})
	return ws.(*websocketImpl)
}

type WebSocket interface {
	Serve(ctx Context, cid uint64, conn *websocket.Conn) error
	Broadcast(ctx stdContext.Context, msg []byte) error
}

type websocketImpl struct {
	ws map[uint64]*websocket.Conn
	mu sync.RWMutex
}

func (w *websocketImpl) Serve(ctx Context, cid uint64, conn *websocket.Conn) error {
	w.mu.Lock()
	if _, ok := w.ws[cid]; ok {
		w.mu.Unlock()
		return errors.New("cid already exists")
	}
	w.ws[cid] = conn
	w.mu.Unlock()

	w.serve(ctx, conn)

	w.mu.Lock()
	delete(w.ws, cid)
	w.mu.Unlock()

	conn.Close(websocket.StatusNormalClosure, "normal closure")

	return nil
}

func (w *websocketImpl) Broadcast(ctx stdContext.Context, msg []byte) error {
	var list []*websocket.Conn
	w.mu.RLock()
	for _, conn := range w.ws {
		list = append(list, conn)
	}
	w.mu.RUnlock()

	for _, conn := range list {
		err := conn.Write(ctx, websocket.MessageText, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *websocketImpl) serve(ctx Context, conn *websocket.Conn) error {
	s := ctx.s()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := conn.Read(ctx)
			if err != nil {
				return err
			}

			w := &websocketResponseWriter{ctx, conn}
			r, err := newWebSocketRequest(ctx, msg)
			if err != nil {
				c := s.AcquireContext()
				defer s.ReleaseContext(c)
				c.SetRequest(r)
				c.SetResponseWriter(NewResponseWriter(w))
				c.Error(err)
				s.Logger.Printf("[msg: new websocket request error] [req msg: %s] [err: %v]", string(msg), err)
				continue
			}

			s.ServeHTTP(w, r)
		}
	}
}

type websocketResponseWriter struct {
	ctx  Context
	conn *websocket.Conn
}

func (w *websocketResponseWriter) Write(p []byte) (n int, err error) {
	err = w.conn.Write(w.ctx, websocket.MessageText, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *websocketResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *websocketResponseWriter) WriteHeader(int) {}

func (w *websocketResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

type websocketRequest struct {
	RID    uint64            `json:"rid"`
	URI    string            `json:"uri"`
	Header map[string]string `json:"header"`
	Body   json.RawMessage   `json:"body"`
}

func newWebSocketRequest(ctx Context, msg []byte) (*http.Request, error) {
	var r websocketRequest
	err := json.Unmarshal(msg, &r)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if r.RID == 0 {
		return nil, fmt.Errorf("rid is required")
	}

	if r.URI == "" {
		return nil, fmt.Errorf("uri is required")
	}
	uri, err := url.Parse(r.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid uri: %w", err)
	}

	req := ctx.Request().Clone(ctx)
	for k, v := range r.Header {
		req.Header.Set(k, v)
	}
	req.Method = MethodWebSocket
	req.RequestURI = uri.RequestURI()
	req.URL.Path = uri.Path
	req.Header.Set(HeaderXRequestID, strconv.FormatUint(r.RID, 10))
	req.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	req.Body = io.NopCloser(bytes.NewReader(r.Body))
	req.ContentLength = int64(len(r.Body))

	return req, nil
}
