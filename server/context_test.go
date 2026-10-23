package server

import (
	"bufio"
	"errors"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEmbedStruct struct {
	Name string `json:"name" bind:"query,form"`
}

type testBindStruct struct {
	UID     int    `json:"uid" bind:"ctx"`
	Name    string `json:"name" bind:"query,form"`
	Age     int    `json:"age" bind:"query,form"`
	Email   string `json:"email" bind:"header"`
	Session string `json:"session" bind:"cookie"`
	ID      string `json:"id" bind:"param"`
	IDs     []int  `json:"ids" bind:"form"`
}

type testNestedStruct struct {
	Basic    testBindStruct  `json:"basic" bind:"form"`
	Embed    testEmbedStruct `json:"embed" bind:"form"`
	Pointer  *testBindStruct `json:"pointer" bind:"form"`
	NoTag    testBindStruct
	NoExport testBindStruct `json:"-"`
}

func TestContextBind(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		setup       func(*context)
		input       string
		want        testBindStruct
		wantErr     bool
	}{
		{
			name:        "bind json",
			contentType: MIMEApplicationJSON,
			setup:       func(c *context) {},
			input:       `{"uid":12, "name":"test","age":18}`,
			want: testBindStruct{
				UID:  0,
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind form",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"age":  {"18"},
				}
			},
			want: testBindStruct{
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind form array",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"ids":  {"18,19,20"},
				}
			},
			want: testBindStruct{
				Name: "test",
				IDs:  []int{18, 19, 20},
			},
		},
		{
			name:        "bind form json",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"name": {"test"},
					"ids":  {`[18,19,20]`},
				}
			},
			want: testBindStruct{
				Name: "test",
				IDs:  []int{18, 19, 20},
			},
		},
		{
			name:        "bind query",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.URL.RawQuery = "name=test&age=18"
			},
			want: testBindStruct{
				Name: "test",
				Age:  18,
			},
		},
		{
			name:        "bind header",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Header.Set("email", "test@example.com")
			},
			want: testBindStruct{
				Email: "test@example.com",
			},
		},
		{
			name:        "bind cookie",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.AddCookie(&http.Cookie{Name: "session", Value: "123456"})
			},
			want: testBindStruct{
				Session: "123456",
			},
		},
		{
			name:        "bind param",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.pnames = []string{"id"}
				c.pvalues = []string{"abc123"}
			},
			want: testBindStruct{
				ID: "abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
			r.Header.Set(HeaderContentType, tt.contentType)

			c := &context{
				request:        r,
				responseWriter: NewResponseWriter(w),
				server:         New(),
			}

			tt.setup(c)

			var got testBindStruct
			err := c.Bind(&got)

			if (err != nil) != tt.wantErr {
				t.Errorf("context.Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("context.Bind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextBindNested(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		setup       func(*context)
		input       string
		want        testNestedStruct
		wantErr     bool
	}{
		{
			name:        "bind nested json",
			contentType: MIMEApplicationJSON,
			setup:       func(c *context) {},
			input:       `{"name":"li","basic":{"name":"test","age":18},"embed":{"name":"embed"},"pointer":{"name":"ptr","age":20}}`,
			want: testNestedStruct{
				Basic: testBindStruct{
					Name: "test",
					Age:  18,
				},
				Embed: testEmbedStruct{
					Name: "embed",
				},
				NoTag: testBindStruct{
					Name: "li",
				},
				Pointer: &testBindStruct{
					Name: "ptr",
					Age:  20,
				},
			},
		},
		{
			name:        "bind nested form",
			contentType: MIMEApplicationForm,
			setup: func(c *context) {
				c.request.Form = map[string][]string{
					"basic":   {`{"name":"test","age":18}`},
					"embed":   {`{"name":"embed"}`},
					"pointer": {`{"name":"ptr","age":20}`},
				}
			},
			want: testNestedStruct{
				Basic: testBindStruct{
					Name: "test",
					Age:  18,
				},
				Embed: testEmbedStruct{
					Name: "embed",
				},
				Pointer: &testBindStruct{
					Name: "ptr",
					Age:  20,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
			r.Header.Set(HeaderContentType, tt.contentType)

			c := &context{
				request:        r,
				responseWriter: NewResponseWriter(w),
				server:         New(),
			}

			tt.setup(c)

			var got testNestedStruct
			err := c.Bind(&got)

			if (err != nil) != tt.wantErr {
				t.Errorf("context.Bind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got.Basic, tt.want.Basic) {
				t.Errorf("context.Bind() Basic = %v, want %v", got.Basic, tt.want.Basic)
			}
			if got.Embed != tt.want.Embed {
				t.Errorf("context.Bind() Embed = %v, want %v", got.Embed, tt.want.Embed)
			}
			if got.Pointer == nil || !reflect.DeepEqual(*got.Pointer, *tt.want.Pointer) {
				t.Errorf("context.Bind() Pointer = %v, want %v", got.Pointer, tt.want.Pointer)
			}
		})
	}
}

type hrw struct {
}

func (h *hrw) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (h *hrw) WriteHeader(statusCode int) {
}

func (h *hrw) Header() http.Header {
	return http.Header{}
}

func (h *hrw) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (h *hrw) Flush() {
}

func TestContextResponseWriter(t *testing.T) {

	resp := NewResponseWriter(&hrw{})
	resp.Write([]byte("test"))
	resp.WriteHeader(http.StatusOK)
	resp.Flush()
	resp.Hijack()
	resp.Write([]byte("test"))
	resp.WriteHeader(http.StatusOK)
	resp.Flush()
	resp.Hijack()
}

type ctxTestUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

const ctxTestUserJSON = `{"id":1,"name":"Jon Snow"}`

func TestContextJSON(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// JSON
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)
	err := c.JSON(http.StatusOK, ctxTestUser{1, "Jon Snow"})
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(ctxTestUserJSON+"\n", rec.Body.String())
	}

	// JSON (custom code) —— lazygo JSON() 先写 Status 字段、首次 Write 时才提交
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err = c.JSON(http.StatusCreated, ctxTestUser{1, "Jon Snow"})
	if assert.NoError(err) {
		assert.Equal(http.StatusCreated, rec.Code)
		assert.Equal(MIMEApplicationJSONCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal(ctxTestUserJSON+"\n", rec.Body.String())
	}

	// JSON (error) —— 编码失败不应提交响应
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err = c.JSON(http.StatusOK, make(chan bool))
	assert.Error(err)

	// JSON (NaN) —— 编码失败不应提交响应
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err = c.JSON(http.StatusCreated, map[string]float64{"a": math.NaN()})
	if assert.Error(err) {
		assert.False(c.responseWriter.Committed)
	}
}

func TestContextBlob(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.Blob(http.StatusOK, MIMEOctetStream, []byte("blob"))
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(MIMEOctetStream, rec.Header().Get(HeaderContentType))
		assert.Equal("blob", rec.Body.String())
	}
}

func TestContextHTML(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.HTML(http.StatusOK, "Hello, <strong>World!</strong>")
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal("Hello, <strong>World!</strong>", rec.Body.String())
	}
}

func TestContextHTMLBlob(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.HTMLBlob(http.StatusOK, []byte("Hello, <strong>World!</strong>"))
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(MIMETextHTMLCharsetUTF8, rec.Header().Get(HeaderContentType))
		assert.Equal("Hello, <strong>World!</strong>", rec.Body.String())
	}
}

func TestContextStream(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	r := strings.NewReader("response from a stream")
	err := c.Stream(http.StatusOK, MIMEOctetStream, r)
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(MIMEOctetStream, rec.Header().Get(HeaderContentType))
		assert.Equal("response from a stream", rec.Body.String())
	}
}

func TestContextNoContent(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	assert.NoError(c.NoContent(http.StatusNoContent))
	assert.Equal(http.StatusNoContent, rec.Code)
	assert.True(c.responseWriter.Committed)
}

func TestContextRedirect(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	assert.NoError(c.Redirect(http.StatusMovedPermanently, "http://labstack.github.io/echo"))
	assert.Equal(http.StatusMovedPermanently, rec.Code)
	assert.Equal("http://labstack.github.io/echo", rec.Header().Get(HeaderLocation))

	// 非法状态码（不在 300-308 之间）返回 ErrInvalidRedirectCode
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err := c.Redirect(310, "http://labstack.github.io/echo")
	assert.ErrorIs(err, ErrInvalidRedirectCode)
}

func TestContextFile(t *testing.T) {
	assert := assert.New(t)
	s := New()

	// 用 t.TempDir 建临时文件（echo 用 _fixture 目录）
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	assert.NoError(os.WriteFile(file, []byte("file content"), 0o644))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.File(file)
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("file content", rec.Body.String())
		assert.Equal("text/plain; charset=utf-8", rec.Header().Get(HeaderContentType))
	}

	// 文件不存在返回 ErrNotFound
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err = c.File(filepath.Join(dir, "not-exists.txt"))
	assert.ErrorIs(err, ErrNotFound)

	// 目录则查找其中的 index.html
	indexDir := filepath.Join(dir, "subdir")
	assert.NoError(os.Mkdir(indexDir, 0o755))
	assert.NoError(os.WriteFile(filepath.Join(indexDir, "index.html"), []byte("<html>index</html>"), 0o644))
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	err = c.File(indexDir)
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal("<html>index</html>", rec.Body.String())
	}
}

func TestContextAttachment(t *testing.T) {
	assert := assert.New(t)
	s := New()

	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	content := []byte("attachment content")
	assert.NoError(os.WriteFile(file, content, 0o644))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.Attachment(file, "test.txt")
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(`attachment; filename="test.txt"`, rec.Header().Get(HeaderContentDisposition))
		assert.Equal(string(content), rec.Body.String())
	}
}

func TestContextInline(t *testing.T) {
	assert := assert.New(t)
	s := New()

	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	content := []byte("inline content")
	assert.NoError(os.WriteFile(file, content, 0o644))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	err := c.Inline(file, "test.txt")
	if assert.NoError(err) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Equal(`inline; filename="test.txt"`, rec.Header().Get(HeaderContentDisposition))
		assert.Equal(string(content), rec.Body.String())
	}
}

func TestContextCookie(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	theme := "theme=light"
	user := "user=Jon Snow"
	req.Header.Add(HeaderCookie, theme)
	req.Header.Add(HeaderCookie, user)
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)

	// Read single —— lazygo Cookie 返回 (value, ok)
	v, ok := c.Cookie("theme")
	assert.True(ok)
	assert.Equal("light", v)

	v, ok = c.Cookie("user")
	assert.True(ok)
	assert.Equal("Jon Snow", v)

	// 不存在的 cookie
	_, ok = c.Cookie("undefined")
	assert.False(ok)
}

func TestContextParam(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/users/101/files/501", c)

	uid, ok := c.Param("uid")
	assert.True(ok)
	assert.Equal("101", uid)
	fid, ok := c.Param("fid")
	assert.True(ok)
	assert.Equal("501", fid)

	// 未定义的参数
	_, ok = c.Param("undefined")
	assert.False(ok)
}

func TestContextParamValues(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/users/101/files/501", c)

	assert.EqualValues([]string{"101", "501"}, c.ParamValues())
}

func TestContextQueryParam(t *testing.T) {
	assert := assert.New(t)
	q := make(url.Values)
	q.Set("name", "Jon Snow")
	q.Set("email", "jon@labstack.com")
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	s := New()
	c := s.newContext(req, nil).(*context)

	// QueryParam —— lazygo 无 QueryParams()（返回 url.Values 的版本），仅保留 QueryParam 断言
	assert.Equal("Jon Snow", c.QueryParam("name"))
	assert.Equal("jon@labstack.com", c.QueryParam("email"))
}

func TestContextFormValue(t *testing.T) {
	assert := assert.New(t)
	f := make(url.Values)
	f.Set("name", "Jon Snow")
	f.Set("email", "jon@labstack.com")

	s := New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Add(HeaderContentType, MIMEApplicationForm)
	c := s.newContext(req, nil).(*context)

	// FormValue
	assert.Equal("Jon Snow", c.FormValue("name"))
	assert.Equal("jon@labstack.com", c.FormValue("email"))

	// FormParams
	params, err := c.FormParams()
	if assert.NoError(err) {
		assert.Equal(url.Values{
			"name":  []string{"Jon Snow"},
			"email": []string{"jon@labstack.com"},
		}, params)
	}

	// Multipart FormParams error
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	req.Header.Add(HeaderContentType, MIMEMultipartForm)
	c = s.newContext(req, nil).(*context)
	params, err = c.FormParams()
	assert.Nil(params)
	assert.Error(err)
}

func TestContextHandler(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	b := new(strings.Builder)

	r.Add(http.MethodGet, "/handler", func(Context) error {
		_, err := b.WriteString("handler")
		return err
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/handler", c)
	err := c.Handler()(c)
	assert.Equal("handler", b.String())
	assert.NoError(err)
}

func TestContextIsDebug(t *testing.T) {
	assert := assert.New(t)
	s := New()
	c := s.newContext(httptest.NewRequest(http.MethodGet, "/", nil), nil).(*context)

	assert.False(c.IsDebug())

	s.Debug = true
	assert.True(c.IsDebug())
}

func TestContextError(t *testing.T) {
	assert := assert.New(t)
	s := New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// 普通 error 走默认错误处理器 → 500 JSON
	rec := httptest.NewRecorder()
	c := s.newContext(req, rec).(*context)
	c.Error(errors.New("error"))
	assert.Equal(http.StatusInternalServerError, rec.Code)

	// HTTPError 原样使用其状态码
	rec = httptest.NewRecorder()
	c = s.newContext(req, rec).(*context)
	c.Error(ErrNotFound)
	assert.Equal(http.StatusNotFound, rec.Code)
}

func TestContextReset(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	r.Add(http.MethodGet, "/users/:id", func(Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	c := newTestContext(s)
	r.Find(http.MethodGet, "/users/1", c)
	c.WithValue("foe", "ban")
	c.query = url.Values(map[string][]string{"fon": {"baz"}})

	c.reset(req, httptest.NewRecorder())
	assert.Equal(0, len(c.ParamValues()))
	assert.Equal(0, len(c.pnames))
	assert.Nil(c.Value("foe"))
	assert.Equal("", c.GetRoutePath())
	assert.Nil(c.query)
	// func 值不能直接比较，通过执行结果验证 handler 被重置为 NotFoundHandler
	assert.ErrorIs(c.handler(c), ErrNotFound)
}

func TestContextWithValue(t *testing.T) {
	assert := assert.New(t)
	s := New()
	c := s.newContext(httptest.NewRequest(http.MethodGet, "/", nil), nil).(*context)

	c.WithValue("name", "Jon Snow")
	assert.Equal("Jon Snow", c.Value("name"))

	// lazygo 特有语义：写入 nil 删除 key
	c.WithValue("name", nil)
	assert.Nil(c.Value("name"))
}
