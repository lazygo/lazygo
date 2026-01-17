package server

import (
	"bytes"
	stdContext "context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lazygo/lazygo/utils"
)

type (
	File struct {
		File       multipart.File
		FileHeader *multipart.FileHeader
	}

	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		s() *Server

		stdContext.Context
		// Request returns `*http.Request`.
		Request() *http.Request
		SetRequest(*http.Request)

		// ResponseWriter returns `*Response`.
		ResponseWriter() *ResponseWriter
		SetResponseWriter(*ResponseWriter)

		// GetRoutePath route info
		GetRoutePath() string

		Bind(any) error

		// Param returns path parameter by name.
		Param(name string) (string, bool)

		// ParamValues returns path parameter values.
		ParamValues() []string

		GetString(name string, defVal ...string) string
		GetInt(name string, defVal ...int) int
		GetInt64(name string, defVal ...int64) int64
		GetUint(name string, defVal ...uint) uint
		GetUint64(name string, defVal ...uint64) uint64
		PostString(name string, defVal ...string) string
		PostInt(name string, defVal ...int) int
		PostInt64(name string, defVal ...int64) int64
		PostUint(name string, defVal ...uint) uint
		PostUint64(name string, defVal ...uint64) uint64

		Cookie(name string) (string, bool)
		QueryParam(name string) string
		FormValue(name string) string
		FormParams() (url.Values, error)
		FormFile(name string) (*multipart.FileHeader, error)
		MultipartForm() (*multipart.Form, error)

		// WithValue 存入数据到当前请求的context
		WithValue(key string, val any)

		// SetResponseHeader 设置响应头
		SetResponseHeader(headerOptions map[string]string) *context

		// RequestHeader 获取请求头
		RequestHeader(name string) string

		// JSON sends a JSON response with status code.
		JSON(code int, i any) error
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

		// WebSocket 获取websocket对象
		WebSocket(name string) WebSocket

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// IsDebug return the Server is debug.
		IsDebug() bool
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
		server         *Server
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

func (c *context) s() *Server {
	return c.server
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

func (c *context) GetRoutePath() string {
	return c.path
}

func (c *context) Bind(v any) error {
	// result pointer value
	rpv := reflect.ValueOf(v)
	if rpv.Kind() != reflect.Ptr || rpv.IsNil() {
		c.server.Logger.Println("bind value not a pointer")
		return ErrInternalServerError
	}

	req := c.Request()
	ctype := strings.ToLower(req.Header.Get(HeaderContentType))
	if strings.HasPrefix(ctype, MIMEApplicationJSON) && req.ContentLength > 0 {
		buf := &bytes.Buffer{}
		err := json.NewDecoder(io.TeeReader(req.Body, buf)).Decode(v)
		if err != nil {
			return err
		}
		req.Body.Close()
		req.Body = io.NopCloser(buf)
	}

	var fill func(rv reflect.Value) error
	fill = func(rv reflect.Value) error {
		for i := 0; i < rv.NumField(); i++ {
			if !rv.Field(i).CanSet() {
				continue
			}
			tField := rv.Type().Field(i)
			field := tField.Tag.Get("json")
			if field == "" {
				// 如果是嵌套结构体，则递归子结构体进行绑定
				if tField.Type.Kind() == reflect.Pointer && rv.Field(i).IsNil() {
					rv.Field(i).Set(reflect.New(tField.Type.Elem()))
				}

				subrv := reflect.Indirect(rv.Field(i))
				if subrv.Kind() != reflect.Struct {
					continue
				}
				if err := fill(subrv); err != nil {
					return err
				}
				continue
			}

			binds := strings.Split(tField.Tag.Get("bind"), ",")
			procList := strings.Split(tField.Tag.Get("process"), ",")
			var val any
			for _, bind := range binds {
				switch strings.TrimSpace(strings.ToLower(bind)) {
				case "ctx", "context", "value":
					val = c.Value(field)
				case "header":
					val = c.RequestHeader(field)
				case "cookie":
					val, _ = c.Cookie(field)
				case "param":
					val, _ = c.Param(field)
				case "query", "url":
					val = c.QueryParam(field)
				case "form":
					if strings.HasPrefix(ctype, MIMEApplicationForm) || strings.HasPrefix(ctype, MIMEMultipartForm) {
						val = c.FormValue(field)
					}
				case "file":
					if strings.HasPrefix(ctype, MIMEMultipartForm) {
						file, fileHeader, err := req.FormFile(field)
						if err != nil {
							return err
						}
						val = &File{file, fileHeader}
					}
				default:
					continue
				}
				if val != "" && val != nil {
					break
				}
			}

			if len(procList) > 0 && tField.Type.Kind() == reflect.String && !rv.Field(i).IsZero() {
				// 兼容json解析的字段的process处理
				val = rv.Field(i).Interface().(string)
			}
			if val == nil || val == "" {
				if !slices.ContainsFunc(binds, func(value string) bool {
					return slices.Contains([]string{"ctx", "context", "value"}, value)
				}) {
					// 对于context类型的，需要强制覆盖，避免用户通过json请求注入
					continue
				}
			}

			if to, ok := toType(val, tField.Type, procList); ok {
				rv.Field(i).Set(reflect.ValueOf(to))
			}
		}
		return nil
	}

	// result value
	rv := rpv.Elem()
	if rv.Kind() != reflect.Struct {
		c.server.Logger.Println("bind value not a struct pointer")
		return ErrInternalServerError
	}

	return fill(rv)
}

// Param 路由参数
func (c *context) Param(name string) (string, bool) {
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i], true
			}
		}
	}
	return "", false
}

// ParamValues 路由参数
func (c *context) ParamValues() []string {
	return c.pvalues[:len(c.pnames)]
}

// GetString 获取Get字符串变量
func (c *context) GetString(name string, defVal ...string) string {
	return utils.ToString(c.QueryParam(name), defVal...)
}

// GetInt 获取Get整型变量
func (c *context) GetInt(name string, defVal ...int) int {
	return utils.ToInt(c.QueryParam(name), defVal...)
}

// GetInt64 获取Get整型变量
func (c *context) GetInt64(name string, defVal ...int64) int64 {
	return utils.ToInt64(c.QueryParam(name), defVal...)
}

// GetUint 获取Get无符号整型变量
func (c *context) GetUint(name string, defVal ...uint) uint {
	return utils.ToUint(c.QueryParam(name), defVal...)
}

// GetUint64 获取Get无符号整型变量
func (c *context) GetUint64(name string, defVal ...uint64) uint64 {
	return utils.ToUint64(c.QueryParam(name), defVal...)
}

// PostString 获取Post字符串变量
func (c *context) PostString(name string, defVal ...string) string {
	return utils.ToString(c.FormValue(name), defVal...)
}

// PostInt 获取Post整型变量
func (c *context) PostInt(name string, defVal ...int) int {
	return utils.ToInt(c.FormValue(name), defVal...)
}

// PostInt64 获取Post整型变量
func (c *context) PostInt64(name string, defVal ...int64) int64 {
	return utils.ToInt64(c.FormValue(name), defVal...)
}

// PostUint 获取Get无符号整型变量
func (c *context) PostUint(name string, defVal ...uint) uint {
	return utils.ToUint(c.FormValue(name), defVal...)
}

// PostUint64 获取Get无符号整型变量
func (c *context) PostUint64(name string, defVal ...uint64) uint64 {
	return utils.ToUint64(c.FormValue(name), defVal...)
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

func (c *context) WithValue(key string, val any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.store == nil {
		c.store = make(Map)
	}
	if val == nil {
		delete(c.store, key)
		return
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

// RequestHeader 获取请求头
func (c *context) RequestHeader(name string) string {
	return c.request.Header.Get(name)
}

// Cookie cookie
func (c *context) Cookie(name string) (string, bool) {
	val, err := c.request.Cookie(name)
	if err != nil {
		return "", false
	}
	return val.Value, true
}

func (c *context) JSON(code int, i any) error {
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
	c.server.HTTPErrorHandler(err, c)
}

func (c *context) Handler() HandlerFunc {
	return c.handler
}

// WebSocket 获取websocket对象
func (c *context) WebSocket(name string) WebSocket {
	return wsManager.Get(name)
}

func (c *context) IsDebug() bool {
	return c.server.Debug
}

// Deadline returns that there is no deadline (ok==false) when c.Request has no Context.
func (c *context) Deadline() (deadline time.Time, ok bool) {
	return c.request.Context().Deadline()
}

// Done returns nil (chan which will wait forever) when c.Request has no Context.
func (c *context) Done() <-chan struct{} {
	return c.request.Context().Done()
}

// Err returns nil when c.Request has no Context.
func (c *context) Err() error {
	return c.request.Context().Err()
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *context) Value(key any) any {
	if keyAsString, ok := key.(string); ok {
		c.lock.RLock()
		val, ok := c.store[keyAsString]
		c.lock.RUnlock()
		if ok {
			return val
		}
	}
	return c.request.Context().Value(key)
}

// Reset resets the context after request completes. It must be called along
// with `Server#AcquireContext()` and `Server#ReleaseContext()`.
// See `Server#ServeHTTP()`
func (c *context) reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.responseWriter.reset(w)
	c.handler = NotFoundHandler
	c.store = nil
	c.path = ""
	c.pnames = nil
	c.query = nil
	// NOTE: Don't reset because it has to have length c.engine.maxParam at all times
	for i := 0; i < *c.server.maxParam; i++ {
		c.pvalues[i] = ""
	}
}

func toType(val any, rType reflect.Type, procList []string) (any, bool) {
	typeName := rType.String()
	typeKind := rType.Kind()

	if typeKind == reflect.Ptr {
		typeKind = rType.Elem().Kind()
	}
	rv := reflect.ValueOf(val)

	if rv.Kind() == reflect.String && (typeKind == reflect.Array || typeKind == reflect.Interface || typeKind == reflect.Map || typeKind == reflect.Slice || typeKind == reflect.Struct) {
		rVal := reflect.New(rType)
		strVal := utils.ToString(val)
		if err := json.Unmarshal([]byte(strVal), rVal.Interface()); err == nil {
			return rVal.Elem().Interface(), true
		}
	}

	if typeName == "interface {}" {
		return val, true
	}
	switch typeName {
	case "int":
		return utils.ToInt(val), true
	case "[]int":
		returnVal := make([]int, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, utils.ToInt(str))
			}
		}
		return returnVal, true
	case "int8":
		return int8(utils.ToInt(val)), true
	case "[]int8":
		returnVal := make([]int8, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, int8(utils.ToInt(str)))
			}
		}
		return returnVal, true
	case "int16":
		return int16(utils.ToInt(val)), true
	case "[]int16":
		returnVal := make([]int16, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, int16(utils.ToInt(str)))
			}
		}
		return returnVal, true
	case "int32":
		return int32(utils.ToInt(val)), true
	case "[]int32":
		returnVal := make([]int32, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, int32(utils.ToInt(str)))
			}
		}
		return returnVal, true
	case "int64":
		return utils.ToInt64(val), true
	case "[]int64":
		returnVal := make([]int64, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, utils.ToInt64(str))
			}
		}
		return returnVal, true
	case "uint":
		return utils.ToUint(val), true
	case "[]uint":
		returnVal := make([]uint, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, utils.ToUint(str))
			}
		}
		return returnVal, true
	case "uint8":
		return uint8(utils.ToUint(val)), true
	case "[]uint8":
		returnVal := make([]uint8, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, uint8(utils.ToUint(str)))
			}
		}
		return returnVal, true
	case "uint16":
		return uint16(utils.ToUint(val)), true
	case "[]uint16":
		returnVal := make([]uint16, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, uint16(utils.ToUint(str)))
			}
		}
		return returnVal, true
	case "uint32":
		return uint32(utils.ToUint(val)), true
	case "[]uint32":
		returnVal := make([]uint32, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, uint32(utils.ToUint(str)))
			}
		}
		return returnVal, true
	case "uint64":
		return utils.ToUint64(val), true
	case "[]uint64":
		returnVal := make([]uint64, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, utils.ToUint64(str))
			}
		}
		return returnVal, true
	case "float32":
		return float32(utils.ToFloat(val)), true
	case "float64":
		return utils.ToFloat(val), true
	case "string":
		return process(utils.ToString(val), procList), true
	case "[]string":
		returnVal := make([]string, 0)
		strVal := strings.TrimSpace(utils.ToString(val))
		if len(strVal) > 0 {
			for _, str := range strings.Split(strVal, ",") {
				returnVal = append(returnVal, process(utils.ToString(str), procList))
			}
		}
		return returnVal, true
	default:
		valType := rv.Type().String()
		if valType == typeName {
			return val, true
		}
		if strings.HasPrefix(valType, "*") && valType[1:] == typeName {
			val = rv.Elem().Interface()
			return val, true
		}
		return val, false
	}
}

func process(str string, procList []string) string {
	for _, proc := range procList {
		proc = strings.TrimSpace(strings.ToLower(proc))
		switch {
		case strings.HasPrefix(proc, "trim"):
			str = strings.TrimSpace(str)
		case strings.HasPrefix(proc, "tolower"):
			str = strings.ToLower(str)
		case strings.HasPrefix(proc, "toupper"):
			str = strings.ToUpper(str)
		case strings.HasPrefix(proc, "cut("):
			if n, err := strconv.Atoi(proc[4 : len(proc)-1]); err != nil {
				str = utils.CutRune(str, n)
			}
		}
	}
	return str
}
