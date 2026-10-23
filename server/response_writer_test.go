package server

// 移植自 echo v4.1.17 response_test.go，并补充 WriteHeader 幂等、Size 统计、
// reset、非 http.ResponseWriter 底层等用例。
// 适配：lazygo ResponseWriter 无 echo 字段，用 NewResponseWriter(rec) 构造；
// Hijack 在 httptest.NewRecorder 上不支持，改用真实 httptest.NewServer 验证。

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseWriterBeforeAfter(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// Before
	rw.Before(func() {
		rec.Header().Set(HeaderServer, "echo")
	})
	// After
	rw.After(func() {
		rec.Header().Set(HeaderXFrameOptions, "DENY")
	})
	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal("echo", rec.Header().Get(HeaderServer))
	assert.Equal("DENY", rec.Header().Get(HeaderXFrameOptions))
	assert.True(rw.Committed)
	assert.Equal(http.StatusOK, rw.Status)
	assert.Equal(int64(4), rw.Size)
}

func TestResponseWriterWriteHeaderIdempotent(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.WriteHeader(http.StatusCreated)
	rw.WriteHeader(http.StatusOK) // 二次调用不覆盖已提交的状态码
	assert.Equal(http.StatusCreated, rec.Code)
	assert.Equal(http.StatusCreated, rw.Status)
	assert.True(rw.Committed)
}

func TestResponseWriterWriteImplicitOK(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	// 未显式 WriteHeader 时，Write 隐式提交 200
	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal(http.StatusOK, rec.Code)
	assert.Equal(http.StatusOK, rw.Status)
	assert.True(rw.Committed)
}

func TestResponseWriterWriteUsesPresetStatus(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Status = http.StatusBadRequest
	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal(http.StatusBadRequest, rec.Code)
}

func TestResponseWriterFlush(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	rw.Flush()
	assert.True(rec.Flushed)
}

func TestResponseWriterHijack(t *testing.T) {
	assert := assert.New(t)
	s := New()
	s.Get("/hijack", func(c Context) error {
		conn, bufrw, err := c.ResponseWriter().Hijack()
		if err != nil {
			return err
		}
		defer conn.Close()
		// hijack 后需自行写响应行
		_, err = fmt.Fprintf(bufrw, "HTTP/1.1 200 OK\r\nContent-Length: 8\r\n\r\nhijacked")
		assert.NoError(err)
		return bufrw.Flush()
	})

	srv := httptest.NewServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/hijack")
	if assert.NoError(err) {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if assert.NoError(err) {
			assert.Equal("hijacked", string(body))
		}
	}
}

func TestResponseWriterSize(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	n, err := rw.Write([]byte("foo"))
	assert.NoError(err)
	assert.Equal(3, n)
	assert.Equal(int64(3), rw.Size)

	n, err = rw.Write([]byte("bar"))
	assert.NoError(err)
	assert.Equal(3, n)
	assert.Equal(int64(6), rw.Size)
	assert.Equal("foobar", rec.Body.String())
}

func TestResponseWriterReset(t *testing.T) {
	assert := assert.New(t)
	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)

	rw.Before(func() {})
	rw.After(func() {})
	rw.WriteHeader(http.StatusCreated)
	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	assert.True(rw.Committed)

	rec2 := httptest.NewRecorder()
	rw.reset(rec2)
	assert.Equal(rec2, rw.Writer)
	assert.Equal(int64(0), rw.Size)
	assert.Equal(http.StatusOK, rw.Status)
	assert.False(rw.Committed)
	assert.Empty(rw.beforeFuncs)
	assert.Empty(rw.afterFuncs)

	// reset 后可正常再次使用
	_, err = rw.Write([]byte("again"))
	assert.NoError(err)
	assert.Equal("again", rec2.Body.String())
	assert.Equal(int64(5), rw.Size)
}

func TestResponseWriterPlainIOWriter(t *testing.T) {
	assert := assert.New(t)
	// 底层不是 http.ResponseWriter（如 Server.NewIOWriterContext 场景）
	buf := new(bytes.Buffer)
	rw := NewResponseWriter(buf)

	// Header 返回空 header，不 panic
	assert.NotNil(rw.Header())
	assert.Empty(rw.Header())

	// WriteHeader 静默处理，不 panic；Status/Committed 照常记录
	rw.WriteHeader(http.StatusOK)
	assert.True(rw.Committed)
	assert.Equal(http.StatusOK, rw.Status)

	_, err := rw.Write([]byte("test"))
	assert.NoError(err)
	assert.Equal(int64(4), rw.Size)
	assert.Equal("test", buf.String())

	// Unwrap 返回 nil（已在 regress_test 中验证 http.ResponseWriter 场景）
	assert.Nil(rw.Unwrap())

	// Hijack 报错但不 panic
	_, _, err = rw.Hijack()
	assert.Error(err)

	// Flush 不 panic
	rw.Flush()
}
