package engine

import (
	"encoding/json"
	"fmt"
	"github.com/lazygo/lazygo/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		// Request returns `*http.Request`.
		Request() *http.Request
		SetRequest(*http.Request)

		// ResponseWriter returns `*Response`.
		ResponseWriter() *ResponseWriter
		SetResponseWriter(*ResponseWriter)

		// Param returns path parameter by name.
		Param(name string) string

		// ParamValues returns path parameter values.
		ParamValues() []string

		GetString(name string, defVal ...string) string
		GetInt(name string, defVal ...int) int
		GetInt64(name string, defVal ...int64) int64
		PostString(name string, defVal ...string) string
		PostInt(name string, defVal ...int) int
		PostInt64(name string, defVal ...int64) int64

		QueryParam(name string) string
		FormValue(name string) string
		FormParams() (url.Values, error)
		FormFile(name string) (*multipart.FileHeader, error)
		MultipartForm() (*multipart.Form, error)

		// GetVar 取出存入当前请求context的数据
		GetVar(key string) interface{}

		// SetVar 存入数据到当前请求的context
		SetVar(key string, val interface{})

		// SetResponseHeader 设置响应头
		SetResponseHeader(headerOptions map[string]string) *context

		// GetRequestHeader 获取请求头
		GetRequestHeader(name string) string

		// 失败响应
		ApiFail(code int, message string, data interface{}) error
		// 成功响应
		ApiSucc(data map[string]interface{}, message string) error

		// JSON sends a JSON response with status code.
		JSON(code int, i interface{}) error
		// Blob sends a blob response with status code and content type.
		Blob(code int, contentType string, b []byte) error

		// HTML sends an HTTP response with status code.
		HTML(code int, html string) error
		// HTMLBlob sends an HTTP blob response with status code.
		HTMLBlob(code int, b []byte) error

		// Stream sends a streaming response with status code and content type.
		Stream(code int, contentType string, r io.Reader) error
		// File sends a response with the content of the file.
		File(file string) error

		// Attachment sends a response as attachment, prompting client to save the
		// file.
		Attachment(file string, name string) error

		// Inline sends a response as inline, opening the file in the browser.
		Inline(file string, name string) error

		// NoContent sends a response with no body and a status code.
		NoContent(code int) error

		// Redirect redirects the request to a provided URL with status code.
		Redirect(code int, url string) error

		// Error invokes the registered HTTP error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc
	}

	context struct {
		request        *http.Request
		responseWriter *ResponseWriter
		path           string
		pnames         []string
		pvalues        []string
		query          url.Values
		handler        HandlerFunc
		store          Map
		engine         *Engine
		lock           sync.RWMutex
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
	defaultIndent = "  "
)

func (c *context) writeContentType(value string) {
	header := c.responseWriter.Header()
	if header.Get(HeaderContentType) == "" {
		header.Set(HeaderContentType, value)
	}
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) SetRequest(r *http.Request) {
	c.request = r
}

func (c *context) ResponseWriter() *ResponseWriter {
	return c.responseWriter
}

func (c *context) SetResponseWriter(w *ResponseWriter) {
	c.responseWriter = w
}

// 路由参数
func (c *context) Param(name string) string {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

// 路由参数
func (c *context) ParamValues() []string {
	return c.pvalues[:len(c.pnames)]
}

// 获取Get字符串变量
func (c *context) GetString(name string, defVal ...string) string {
	return utils.ToString(c.QueryParam(name), defVal...)
}

// 获取Get整型变量
func (c *context) GetInt(name string, defVal ...int) int {
	return utils.ToInt(c.QueryParam(name), defVal...)
}

// 获取Get整型变量
func (c *context) GetInt64(name string, defVal ...int64) int64 {
	return utils.ToInt64(c.QueryParam(name), defVal...)
}

// 获取Post字符串变量
func (c *context) PostString(name string, defVal ...string) string {
	return utils.ToString(c.FormValue(name), defVal...)
}

// 获取Post整型变量
func (c *context) PostInt(name string, defVal ...int) int {
	return utils.ToInt(c.FormValue(name), defVal...)
}

// 获取Post整型变量
func (c *context) PostInt64(name string, defVal ...int64) int64 {
	return utils.ToInt64(c.FormValue(name), defVal...)
}

func (c *context) QueryParam(name string) string {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query.Get(name)
}

func (c *context) FormValue(name string) string {
	return c.request.FormValue(name)
}

func (c *context) FormParams() (url.Values, error) {
	if strings.HasPrefix(c.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := c.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return c.request.Form, nil
}

func (c *context) FormFile(name string) (*multipart.FileHeader, error) {
	f, fh, err := c.request.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, nil
}

func (c *context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

func (c *context) GetVar(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

func (c *context) SetVar(key string, val interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = val
}

// SetResponseHeader 设置响应头
func (c *context) SetResponseHeader(headerOptions map[string]string) *context {
	if len(headerOptions) > 0 {
		for field, val := range headerOptions {
			c.responseWriter.Header().Set(field, val)
		}
	}
	return c
}

// GetRequestHeader 获取请求头
func (c *context) GetRequestHeader(name string) string {
	return c.request.Header.Get(name)
}

// 成功响应
func (c *context) ApiSucc(data map[string]interface{}, message string) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	result := map[string]interface{}{
		"code":    200,
		"message": message,
		"data":    data,
	}
	return c.JSON(200, result)
}

// 失败响应
func (c *context) ApiFail(code int, message string, data interface{}) error {
	result := map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	}
	return c.JSON(200, result)
}

func (c *context) JSON(code int, i interface{}) error {
	enc := json.NewEncoder(c.responseWriter)
	c.writeContentType(MIMEApplicationJSONCharsetUTF8)
	c.responseWriter.Status = code
	return enc.Encode(i)
}

func (c *context) Blob(code int, contentType string, b []byte) error {
	c.writeContentType(contentType)
	c.responseWriter.WriteHeader(code)
	_, err := c.responseWriter.Write(b)
	return err
}

func (c *context) Stream(code int, contentType string, r io.Reader) error {
	c.writeContentType(contentType)
	c.responseWriter.WriteHeader(code)
	_, err := io.Copy(c.responseWriter, r)
	return err
}

func (c *context) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return err
		}
	}
	http.ServeContent(c.responseWriter, c.Request(), fi.Name(), fi.ModTime(), f)
	return err
}

func (c *context) HTML(code int, html string) error {
	return c.HTMLBlob(code, []byte(html))
}

func (c *context) HTMLBlob(code int, b []byte) error {
	return c.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (c *context) Attachment(file, name string) error {
	return c.contentDisposition(file, name, "attachment")
}

func (c *context) Inline(file, name string) error {
	return c.contentDisposition(file, name, "inline")
}

func (c *context) contentDisposition(file, name, dispositionType string) error {
	c.responseWriter.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	return c.File(file)
}

func (c *context) NoContent(code int) error {
	c.responseWriter.WriteHeader(code)
	return nil
}

func (c *context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	c.responseWriter.Header().Set(HeaderLocation, url)
	c.responseWriter.WriteHeader(code)
	return nil
}

func (c *context) Error(err error) {
	c.engine.HTTPErrorHandler(err, c)
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

// Reset resets the context after request completes. It must be called along
// with `Engine#AcquireContext()` and `Engine#ReleaseContext()`.
// See `Engine#ServeHTTP()`
func (c *context) reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.responseWriter.reset(w)
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil
	c.query = nil
	// NOTE: Don't reset because it has to have length c.engine.maxParam at all times
	for i := 0; i < *c.engine.maxParam; i++ {
		c.pvalues[i] = ""
	}
}
