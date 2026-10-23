package server

// server包缺陷修复的回归测试

import (
	"net/http"
	"net/http/httptest"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

// issue #1653：param节点只有子路由没有根处理器时，Find失配路径会pvalues[-1]越界panic
func TestRouterParamNoRootOnlyChilds(t *testing.T) {
	assert := testify.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/users/create", func(c Context) error { return nil })
	r.Add(http.MethodGet, "/users/:id/edit", func(c Context) error { return nil })
	r.Add(http.MethodGet, "/users/:id/active", func(c Context) error { return nil })

	c := s.newContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/users/alice/edit", c)
	id, ok := c.Param("id")
	assert.True(ok)
	assert.Equal("alice", id)

	// 修复前：该路径触发panic（pvalues[-1]越界）
	c = s.newContext(nil, nil).(*context)
	r.Find(http.MethodGet, "/users/createNotFound", c)
	err := c.handler(c)
	he, ok := err.(*HTTPError)
	assert.True(ok)
	assert.Equal(http.StatusNotFound, he.Code)
}

// issue #1739/#1741：部分前缀失配且带尾斜杠时，不应错误落入any路由
func TestRouterMatchAnyPartialPrefix(t *testing.T) {
	assert := testify.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/*", func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	})
	r.Add(http.MethodGet, "/users/*", func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	})

	find := func(path string) (string, string) {
		c := s.newContext(nil, nil).(*context)
		r.Find(http.MethodGet, path, c)
		assert.Nil(c.handler(c))
		p, _ := c.Param("*")
		return c.Value("path").(string), p
	}

	path, star := find("/users")
	assert.Equal("/*", path)
	assert.Equal("users", star)

	path, star = find("/users/")
	assert.Equal("/users/*", path)
	assert.Equal("", star)

	// /users_prefix/ 只有部分前缀匹配 /users/*，应落入 /* 且保留完整剩余路径
	path, star = find("/users_prefix")
	assert.Equal("/*", path)
	assert.Equal("users_prefix", star)

	path, star = find("/users_prefix/")
	assert.Equal("/*", path)
	assert.Equal("users_prefix/", star)
}

// echo #3157：context创建后注册更多参数的路由，pvalues需按需扩容
func TestPvaluesGrowAfterContextCreated(t *testing.T) {
	assert := testify.New(t)
	s := New()

	// 模拟context先于路由创建（连接池预热/AcquireContext）
	c := s.newContext(nil, nil).(*context)
	assert.Equal(0, len(c.pvalues))

	s.Get("/a/:x/:y/:z", func(c Context) error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/a/1/2/3", nil)
	rec := httptest.NewRecorder()
	// 修复前：reset中 c.pvalues[i] = "" 越界panic
	c.reset(req, rec)
	s.router.Find(http.MethodGet, "/a/1/2/3", c)
	v, ok := c.Param("x")
	assert.True(ok)
	assert.Equal("1", v)
	v, ok = c.Param("z")
	assert.True(ok)
	assert.Equal("3", v)
}

// process cut()：修复前条件写反导致cut永不生效，非法参数反而截断为空串
func TestBindProcessCut(t *testing.T) {
	assert := testify.New(t)

	type req struct {
		Name string `json:"name" bind:"query" process:"cut(3)"`
	}
	s := New()
	c := s.newContext(httptest.NewRequest(http.MethodGet, "/?name=abcdef", nil), httptest.NewRecorder()).(*context)
	var r req
	assert.Nil(c.Bind(&r))
	assert.Equal("abc", r.Name)

	// 非法cut参数不截断
	type reqBad struct {
		Name string `json:"name" bind:"query" process:"cut(x)"`
	}
	c = s.newContext(httptest.NewRequest(http.MethodGet, "/?name=abcdef", nil), httptest.NewRecorder()).(*context)
	var rb reqBad
	assert.Nil(c.Bind(&rb))
	assert.Equal("abcdef", rb.Name)
}

// bool字段绑定：修复前toType无bool分支，bool字段永远绑不上
func TestBindBool(t *testing.T) {
	assert := testify.New(t)

	type req struct {
		Flag    bool `json:"flag" bind:"query"`
		Invalid bool `json:"invalid" bind:"query"`
	}
	s := New()
	c := s.newContext(httptest.NewRequest(http.MethodGet, "/?flag=true&invalid=abc", nil), httptest.NewRecorder()).(*context)
	var r req
	assert.Nil(c.Bind(&r))
	assert.True(r.Flag)
	// 非法值不覆盖字段默认值
	assert.False(r.Invalid)
}

// handler panic时统一返回500，而不是断连
func TestPanicReturns500(t *testing.T) {
	assert := testify.New(t)
	s := New()
	s.Get("/panic", func(c Context) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	assert.Equal(http.StatusInternalServerError, rec.Code)
	assert.Contains(rec.Body.String(), "500")
}

// Flush先提交状态；Unwrap穿透包装writer
func TestResponseWriterFlushAndUnwrap(t *testing.T) {
	assert := testify.New(t)

	rec := httptest.NewRecorder()
	rw := NewResponseWriter(rec)
	rw.Flush()
	assert.True(rw.Committed)
	assert.Equal(http.StatusOK, rw.Status)
	assert.Equal(rec, rw.Unwrap())
}
