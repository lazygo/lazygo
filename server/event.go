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
	"strconv"
	"sync"
)

type EventManager struct {
	event  sync.Map
	server *Server
}

func (em *EventManager) Get(method, subject string) *Event {
	e, _ := em.event.LoadOrStore(fmt.Sprintf("%s:%s", method, subject), &Event{
		server:  em.server,
		method:  method,
		subject: subject,
		rwc:     make(map[uint64]io.ReadWriteCloser),
	})
	return e.(*Event)
}

type Event struct {
	mu      sync.RWMutex
	server  *Server
	method  string
	subject string
	rwc     map[uint64]io.ReadWriteCloser
}

func (e *Event) Serve(ctx stdContext.Context, cid uint64, rwc io.ReadWriteCloser) error {
	e.mu.Lock()
	if _, ok := e.rwc[cid]; ok {
		e.mu.Unlock()
		return errors.New("cid already exists")
	}
	e.rwc[cid] = rwc
	e.mu.Unlock()

	e.serve(ctx, rwc)

	e.mu.Lock()
	delete(e.rwc, cid)
	e.mu.Unlock()

	err := rwc.Close()
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) Broadcast(ctx stdContext.Context, msg []byte) error {
	var list []io.ReadWriteCloser
	e.mu.RLock()
	for _, rwc := range e.rwc {
		list = append(list, rwc)
	}
	e.mu.RUnlock()

	for _, rwc := range list {
		_, err := rwc.Write(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Event) serve(ctx stdContext.Context, rwc io.ReadWriteCloser) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 使用 buffer 收集完整响应，避免 ServeHTTP 多次 Write 被拆成多条 WebSocket 消息
			buf := &bytes.Buffer{}
			w := &eventResponseWriter{ctx: ctx, Writer: buf, header: http.Header{}}
			r, err := newEventRequest(ctx, e.method, rwc)
			if err != nil {
				c := e.server.AcquireContext()
				defer e.server.ReleaseContext(c)
				c.SetRequest(r)
				c.SetResponseWriter(NewResponseWriter(w))
				c.Error(err)
				e.server.Logger.Printf("[msg: new websocket request error] [err: %v]", err)
				if buf.Len() > 0 {
					_, _ = rwc.Write(buf.Bytes())
				}
				continue
			}

			e.server.ServeHTTP(w, r)
			// 整份响应作为一条 WebSocket 消息发送，保证不被截断
			if buf.Len() > 0 {
				_, _ = rwc.Write(buf.Bytes())
			}
		}
	}
}

type eventResponseWriter struct {
	ctx stdContext.Context
	io.Writer
	header http.Header
}

func (w *eventResponseWriter) Header() http.Header {
	return w.header
}

func (w *eventResponseWriter) WriteHeader(int) {}

func (w *eventResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

type EventRequest struct {
	RID    uint64            `json:"rid"`
	URI    string            `json:"uri"`
	Header map[string]string `json:"header"`
	Body   json.RawMessage   `json:"body"`
}

func newEventRequest(ctx stdContext.Context, method string, r io.Reader) (*http.Request, error) {
	var e EventRequest
	err := json.NewDecoder(r).Decode(&e)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if e.RID == 0 {
		return nil, fmt.Errorf("rid is required")
	}

	if e.URI == "" {
		return nil, fmt.Errorf("uri is required")
	}

	req, err := http.NewRequestWithContext(ctx, method, e.URI, bytes.NewReader(e.Body))
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	req.ContentLength = int64(len(e.Body))
	for k, v := range e.Header {
		req.Header.Set(k, v)
	}
	req.Header.Set(HeaderXRequestID, strconv.FormatUint(e.RID, 10))
	req.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)

	return req, nil
}
