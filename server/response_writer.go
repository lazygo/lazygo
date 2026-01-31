package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type (
	// ResponseWriter wraps an http.ResponseWriter and implements its interface to be used
	// by an HTTP handler to construct an HTTP response.
	ResponseWriter struct {
		beforeFuncs []func()
		afterFuncs  []func()
		Writer      io.Writer
		Status      int
		Size        int64
		Committed   bool
	}
)

// NewResponseWriter creates a new instance of Response.
func NewResponseWriter(w io.Writer) (r *ResponseWriter) {
	return &ResponseWriter{Writer: w}
}

// Header returns the header map for the writer that will be sent by
// WriteHeader. Changing the header after a call to WriteHeader (or Write) has
// no effect unless the modified headers were declared as trailers by setting
// the "Trailer" header before the call to WriteHeader (see example)
// To suppress implicit response headers, set their value to nil.
func (r *ResponseWriter) Header() http.Header {
	if w, ok := r.Writer.(http.ResponseWriter); ok {
		return w.Header()
	}
	return http.Header{}
}

// Before registers a function which is called just before the response is written.
func (r *ResponseWriter) Before(fn func()) {
	r.beforeFuncs = append(r.beforeFuncs, fn)
}

// After registers a function which is called just after the response is written.
// If the `Content-Length` is unknown, none of the after function is executed.
func (r *ResponseWriter) After(fn func()) {
	r.afterFuncs = append(r.afterFuncs, fn)
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly
// used to send error codes.
func (r *ResponseWriter) WriteHeader(code int) {
	if r.Committed {
		fmt.Println("response already committed")
		return
	}
	for _, fn := range r.beforeFuncs {
		fn()
	}
	r.Status = code
	if w, ok := r.Writer.(http.ResponseWriter); ok {
		w.WriteHeader(code)
	}
	r.Committed = true
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *ResponseWriter) Write(b []byte) (n int, err error) {
	if !r.Committed {
		if r.Status == 0 {
			r.Status = http.StatusOK
		}
		r.WriteHeader(r.Status)
	}
	n, err = r.Writer.Write(b)
	r.Size += int64(n)
	for _, fn := range r.afterFuncs {
		fn()
	}
	return
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
func (r *ResponseWriter) Flush() {
	if flusher, ok := r.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
func (r *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.Writer.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("hijacker not supported")
}

func (r *ResponseWriter) reset(w http.ResponseWriter) {
	r.beforeFuncs = nil
	r.afterFuncs = nil
	r.Writer = w
	r.Size = 0
	r.Status = http.StatusOK
	r.Committed = false
}
