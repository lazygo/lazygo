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

var eventManager = &EventManager{}

type EventManager struct {
	event sync.Map
}

func (e *EventManager) Get(subject string) Event {
	ev, _ := e.event.LoadOrStore(subject, &Event{
		mu:      &sync.RWMutex{},
		subject: subject,
		rwc:     make(map[uint64]io.ReadWriteCloser),
	})
	return ev.(Event)
}

type Event struct {
	mu      *sync.RWMutex
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
			w := &eventResponseWriter{ctx, rwc}
			r, err := newEventRequest(ctx, rwc)
			if err != nil {
				c := s.AcquireContext()
				defer s.ReleaseContext(c)
				c.SetRequest(r)
				c.SetResponseWriter(NewResponseWriter(w))
				c.Error(err)
				s.Logger.Printf("[msg: new websocket request error] [err: %v]", err)
				continue
			}

			s.ServeHTTP(w, r)
		}
	}
}

type eventResponseWriter struct {
	ctx Context
	io.Writer
}

func (w *eventResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *eventResponseWriter) WriteHeader(int) {}

func (w *eventResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("hijack not supported")
}

type eventRequest struct {
	RID    uint64            `json:"rid"`
	URI    string            `json:"uri"`
	Header map[string]string `json:"header"`
	Body   json.RawMessage   `json:"body"`
}

func newEventRequest(ctx Context, r io.Reader) (*http.Request, error) {
	var e eventRequest
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
	req.Method = MethodWebSocket
	req.RequestURI = uri.RequestURI()
	req.URL.Path = uri.Path
	req.Header.Set(HeaderXRequestID, strconv.FormatUint(e.RID, 10))
	req.Header.Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	req.Body = io.NopCloser(bytes.NewReader(e.Body))
	req.ContentLength = int64(len(e.Body))

	return req, nil
}
