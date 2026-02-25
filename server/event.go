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
)

type EventManager struct {
	event sync.Map
}

func (e *EventManager) Get(method, subject string) *Event {
	ev, _ := e.event.LoadOrStore(fmt.Sprintf("%s:%s", method, subject), &Event{
		method:  method,
		subject: subject,
		rwc:     make(map[uint64]io.ReadWriteCloser),
	})
	return ev.(*Event)
}

type Event struct {
	mu      sync.RWMutex
	method  string
	subject string
	rwc     map[uint64]io.ReadWriteCloser
}

func (e *Event) Serve(ctx Context, cid uint64, rwc io.ReadWriteCloser) error {
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

func (e *Event) serve(ctx Context, rwc io.ReadWriteCloser) error {
	s := ctx.s()
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
				c := s.AcquireContext()
				defer s.ReleaseContext(c)
				c.SetRequest(r)
				c.SetResponseWriter(NewResponseWriter(w))
				c.Error(err)
				s.Logger.Printf("[msg: new websocket request error] [err: %v]", err)
				if buf.Len() > 0 {
					_, _ = rwc.Write(buf.Bytes())
				}
				continue
			}

			s.ServeHTTP(w, r)
			// 整份响应作为一条 WebSocket 消息发送，保证不被截断
			if buf.Len() > 0 {
				_, _ = rwc.Write(buf.Bytes())
			}
		}
	}
}

type eventResponseWriter struct {
	ctx Context
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

func newEventRequest(ctx Context, method string, r io.Reader) (*http.Request, error) {
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
	uri, err := url.Parse(e.URI)
	if err != nil {
		return nil, fmt.Errorf("invalid uri: %w", err)
	}

	req := ctx.Request().Clone(ctx)
	for k, v := range e.Header {
		req.Header.Set(k, v)
	}
	req.Method = method
	req.RequestURI = uri.RequestURI()
	req.URL.Path = uri.Path
	req.Header.Set(HeaderXRequestID, strconv.FormatUint(e.RID, 10))
	req.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	req.Body = io.NopCloser(bytes.NewReader(e.Body))
	req.ContentLength = int64(len(e.Body))

	return req, nil
}
