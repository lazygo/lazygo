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
		src:     make(map[uint64]SendReceiveCloser),
	})
	return e.(*Event)
}

type SendReceiveCloser interface {
	Send(data *EventData) error
	Receive(ctx stdContext.Context) (*EventData, error)
	Close() error
}

type Event struct {
	mu      sync.RWMutex
	server  *Server
	method  string
	subject string
	src     map[uint64]SendReceiveCloser
}

func (e *Event) Serve(ctx stdContext.Context, cid uint64, src SendReceiveCloser) error {
	e.mu.Lock()
	if _, ok := e.src[cid]; ok {
		e.mu.Unlock()
		return errors.New("cid already exists")
	}
	e.src[cid] = src
	e.mu.Unlock()

	e.serve(ctx, src)

	e.mu.Lock()
	delete(e.src, cid)
	e.mu.Unlock()

	err := src.Close()
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) Broadcast(ctx stdContext.Context, data *EventData) error {
	var list []SendReceiveCloser
	e.mu.RLock()
	for _, src := range e.src {
		list = append(list, src)
	}
	e.mu.RUnlock()

	var errs error
	for _, src := range list {
		err := src.Send(data)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (e *Event) serve(ctx stdContext.Context, src SendReceiveCloser) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			req, err := src.Receive(ctx)
			if err != nil {
				return err
			}
			if req == nil {
				continue
			}
			// 使用 buffer 收集完整响应，避免 ServeHTTP 多次 Write 被拆成多条 WebSocket 消息
			buf := &bytes.Buffer{}
			w := &eventResponseWriter{ctx: ctx, Writer: buf, header: http.Header{}}
			r, err := newEventRequest(ctx, e.method, req)
			if err != nil {

				c := e.server.AcquireContext()
				c.SetRequest(r)
				c.SetResponseWriter(NewResponseWriter(w))
				c.Error(err)
				e.server.ReleaseContext(c)

				e.server.Logger.Printf("[msg: new websocket request error] [err: %v]", err)
				if buf.Len() > 0 {
					_ = src.Send(&EventData{
						RID:    req.RID,
						URI:    req.URI,
						Header: w.Header(),
						Body:   buf.Bytes(),
					})
				}
				continue
			}

			e.server.ServeHTTP(w, r)
			// 整份响应作为一条 WebSocket 消息发送，保证不被截断
			if buf.Len() > 0 {
				_ = src.Send(&EventData{
					RID:    req.RID,
					URI:    req.URI,
					Header: w.Header(),
					Body:   buf.Bytes(),
				})
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

type EventData struct {
	RID    uint64          `json:"rid"`
	URI    string          `json:"uri"`
	Header http.Header     `json:"header"`
	Body   json.RawMessage `json:"body"`
}

func newEventRequest(ctx stdContext.Context, method string, e *EventData) (*http.Request, error) {
	if e.RID == 0 {
		return nil, fmt.Errorf("rid is required")
	}

	if e.URI == "" {
		return nil, fmt.Errorf("uri is required")
	}

	ctx = stdContext.WithValue(ctx, HeaderXRequestID, e.RID)
	req, err := http.NewRequestWithContext(ctx, method, e.URI, bytes.NewReader(e.Body))
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	req.ContentLength = int64(len(e.Body))
	req.Header = e.Header

	return req, nil
}
