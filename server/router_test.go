package server

// 移植自 echo v4.1.17 router_test.go（lazygo server 包与 echo v4.1.17 路由实现同代）
// API 适配：New()/newContext(r,w).(*context)、Find(method,path,c *context)、
// Param(name) (string,bool)、Get/GetRoutePath、Route 类型冲突改用 testRoute。

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testRoute 对应 echo 的 Route{Method, Path, Name}，lazygo 的 Route 类型已被占用
type testRoute struct {
	Method string
	Path   string
}

var gitHubAPI = []testRoute{
	// OAuth Authorizations
	{"GET", "/authorizations"},
	{"GET", "/authorizations/:id"},
	{"POST", "/authorizations"},
	//{"PUT", "/authorizations/clients/:client_id"},
	//{"PATCH", "/authorizations/:id"},
	{"DELETE", "/authorizations/:id"},
	{"GET", "/applications/:client_id/tokens/:access_token"},
	{"DELETE", "/applications/:client_id/tokens"},
	{"DELETE", "/applications/:client_id/tokens/:access_token"},

	// Activity
	{"GET", "/events"},
	{"GET", "/repos/:owner/:repo/events"},
	{"GET", "/networks/:owner/:repo/events"},
	{"GET", "/orgs/:org/events"},
	{"GET", "/users/:user/received_events"},
	{"GET", "/users/:user/received_events/public"},
	{"GET", "/users/:user/events"},
	{"GET", "/users/:user/events/public"},
	{"GET", "/users/:user/events/orgs/:org"},
	{"GET", "/feeds"},
	{"GET", "/notifications"},
	{"GET", "/repos/:owner/:repo/notifications"},
	{"PUT", "/notifications"},
	{"PUT", "/repos/:owner/:repo/notifications"},
	{"GET", "/notifications/threads/:id"},
	//{"PATCH", "/notifications/threads/:id"},
	{"GET", "/notifications/threads/:id/subscription"},
	{"PUT", "/notifications/threads/:id/subscription"},
	{"DELETE", "/notifications/threads/:id/subscription"},
	{"GET", "/repos/:owner/:repo/stargazers"},
	{"GET", "/users/:user/starred"},
	{"GET", "/user/starred"},
	{"GET", "/user/starred/:owner/:repo"},
	{"PUT", "/user/starred/:owner/:repo"},
	{"DELETE", "/user/starred/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/subscribers"},
	{"GET", "/users/:user/subscriptions"},
	{"GET", "/user/subscriptions"},
	{"GET", "/repos/:owner/:repo/subscription"},
	{"PUT", "/repos/:owner/:repo/subscription"},
	{"DELETE", "/repos/:owner/:repo/subscription"},
	{"GET", "/user/subscriptions/:owner/:repo"},
	{"PUT", "/user/subscriptions/:owner/:repo"},
	{"DELETE", "/user/subscriptions/:owner/:repo"},

	// Gists
	{"GET", "/users/:user/gists"},
	{"GET", "/gists"},
	//{"GET", "/gists/public"},
	//{"GET", "/gists/starred"},
	{"GET", "/gists/:id"},
	{"POST", "/gists"},
	//{"PATCH", "/gists/:id"},
	{"PUT", "/gists/:id/star"},
	{"DELETE", "/gists/:id/star"},
	{"GET", "/gists/:id/star"},
	{"POST", "/gists/:id/forks"},
	{"DELETE", "/gists/:id"},

	// Git Data
	{"GET", "/repos/:owner/:repo/git/blobs/:sha"},
	{"POST", "/repos/:owner/:repo/git/blobs"},
	{"GET", "/repos/:owner/:repo/git/commits/:sha"},
	{"POST", "/repos/:owner/:repo/git/commits"},
	//{"GET", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/refs"},
	{"POST", "/repos/:owner/:repo/git/refs"},
	//{"PATCH", "/repos/:owner/:repo/git/refs/*ref"},
	//{"DELETE", "/repos/:owner/:repo/git/refs/*ref"},
	{"GET", "/repos/:owner/:repo/git/tags/:sha"},
	{"POST", "/repos/:owner/:repo/git/tags"},
	{"GET", "/repos/:owner/:repo/git/trees/:sha"},
	{"POST", "/repos/:owner/:repo/git/trees"},

	// Issues
	{"GET", "/issues"},
	{"GET", "/user/issues"},
	{"GET", "/orgs/:org/issues"},
	{"GET", "/repos/:owner/:repo/issues"},
	{"GET", "/repos/:owner/:repo/issues/:number"},
	{"POST", "/repos/:owner/:repo/issues"},
	//{"PATCH", "/repos/:owner/:repo/issues/:number"},
	{"GET", "/repos/:owner/:repo/assignees"},
	{"GET", "/repos/:owner/:repo/assignees/:assignee"},
	{"GET", "/repos/:owner/:repo/issues/:number/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments"},
	//{"GET", "/repos/:owner/:repo/issues/comments/:id"},
	{"POST", "/repos/:owner/:repo/issues/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/issues/comments/:id"},
	//{"DELETE", "/repos/:owner/:repo/issues/comments/:id"},
	{"GET", "/repos/:owner/:repo/issues/:number/events"},
	//{"GET", "/repos/:owner/:repo/issues/events"},
	//{"GET", "/repos/:owner/:repo/issues/events/:id"},
	{"GET", "/repos/:owner/:repo/labels"},
	{"GET", "/repos/:owner/:repo/labels/:name"},
	{"POST", "/repos/:owner/:repo/labels"},
	//{"PATCH", "/repos/:owner/:repo/labels/:name"},
	{"DELETE", "/repos/:owner/:repo/labels/:name"},
	{"GET", "/repos/:owner/:repo/issues/:number/labels"},
	{"POST", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels/:name"},
	{"PUT", "/repos/:owner/:repo/issues/:number/labels"},
	{"DELETE", "/repos/:owner/:repo/issues/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones/:number/labels"},
	{"GET", "/repos/:owner/:repo/milestones"},
	{"GET", "/repos/:owner/:repo/milestones/:number"},
	{"POST", "/repos/:owner/:repo/milestones"},
	//{"PATCH", "/repos/:owner/:repo/milestones/:number"},
	{"DELETE", "/repos/:owner/:repo/milestones/:number"},

	// Miscellaneous
	{"GET", "/emojis"},
	{"GET", "/gitignore/templates"},
	{"GET", "/gitignore/templates/:name"},
	{"POST", "/markdown"},
	{"POST", "/markdown/raw"},
	{"GET", "/meta"},
	{"GET", "/rate_limit"},

	// Organizations
	{"GET", "/users/:user/orgs"},
	{"GET", "/user/orgs"},
	{"GET", "/orgs/:org"},
	//{"PATCH", "/orgs/:org"},
	{"GET", "/orgs/:org/members"},
	{"GET", "/orgs/:org/members/:user"},
	{"DELETE", "/orgs/:org/members/:user"},
	{"GET", "/orgs/:org/public_members"},
	{"GET", "/orgs/:org/public_members/:user"},
	{"PUT", "/orgs/:org/public_members/:user"},
	{"DELETE", "/orgs/:org/public_members/:user"},
	{"GET", "/orgs/:org/teams"},
	{"GET", "/teams/:id"},
	{"POST", "/orgs/:org/teams"},
	//{"PATCH", "/teams/:id"},
	{"DELETE", "/teams/:id"},
	{"GET", "/teams/:id/members"},
	{"GET", "/teams/:id/members/:user"},
	{"PUT", "/teams/:id/members/:user"},
	{"DELETE", "/teams/:id/members/:user"},
	{"GET", "/teams/:id/repos"},
	{"GET", "/teams/:id/repos/:owner/:repo"},
	{"PUT", "/teams/:id/repos/:owner/:repo"},
	{"DELETE", "/teams/:id/repos/:owner/:repo"},
	{"GET", "/user/teams"},

	// Pull Requests
	{"GET", "/repos/:owner/:repo/pulls"},
	{"GET", "/repos/:owner/:repo/pulls/:number"},
	{"POST", "/repos/:owner/:repo/pulls"},
	//{"PATCH", "/repos/:owner/:repo/pulls/:number"},
	{"GET", "/repos/:owner/:repo/pulls/:number/commits"},
	{"GET", "/repos/:owner/:repo/pulls/:number/files"},
	{"GET", "/repos/:owner/:repo/pulls/:number/merge"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/merge"},
	{"GET", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments"},
	//{"GET", "/repos/:owner/:repo/pulls/comments/:number"},
	{"PUT", "/repos/:owner/:repo/pulls/:number/comments"},
	//{"PATCH", "/repos/:owner/:repo/pulls/comments/:number"},
	//{"DELETE", "/repos/:owner/:repo/pulls/comments/:number"},

	// Repositories
	{"GET", "/user/repos"},
	{"GET", "/users/:user/repos"},
	{"GET", "/orgs/:org/repos"},
	{"GET", "/repositories"},
	{"POST", "/user/repos"},
	{"POST", "/orgs/:org/repos"},
	{"GET", "/repos/:owner/:repo"},
	//{"PATCH", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/contributors"},
	{"GET", "/repos/:owner/:repo/languages"},
	{"GET", "/repos/:owner/:repo/teams"},
	{"GET", "/repos/:owner/:repo/tags"},
	{"GET", "/repos/:owner/:repo/branches"},
	{"GET", "/repos/:owner/:repo/branches/:branch"},
	{"DELETE", "/repos/:owner/:repo"},
	{"GET", "/repos/:owner/:repo/collaborators"},
	{"GET", "/repos/:owner/:repo/collaborators/:user"},
	{"PUT", "/repos/:owner/:repo/collaborators/:user"},
	{"DELETE", "/repos/:owner/:repo/collaborators/:user"},
	{"GET", "/repos/:owner/:repo/comments"},
	{"GET", "/repos/:owner/:repo/commits/:sha/comments"},
	{"POST", "/repos/:owner/:repo/commits/:sha/comments"},
	{"GET", "/repos/:owner/:repo/comments/:id"},
	//{"PATCH", "/repos/:owner/:repo/comments/:id"},
	{"DELETE", "/repos/:owner/:repo/comments/:id"},
	{"GET", "/repos/:owner/:repo/commits"},
	{"GET", "/repos/:owner/:repo/commits/:sha"},
	{"GET", "/repos/:owner/:repo/readme"},
	//{"GET", "/repos/:owner/:repo/contents/*path"},
	//{"PUT", "/repos/:owner/:repo/contents/*path"},
	//{"DELETE", "/repos/:owner/:repo/contents/*path"},
	//{"GET", "/repos/:owner/:repo/:archive_format/:ref"},
	{"GET", "/repos/:owner/:repo/keys"},
	{"GET", "/repos/:owner/:repo/keys/:id"},
	{"POST", "/repos/:owner/:repo/keys"},
	//{"PATCH", "/repos/:owner/:repo/keys/:id"},
	{"DELETE", "/repos/:owner/:repo/keys/:id"},
	{"GET", "/repos/:owner/:repo/downloads"},
	{"GET", "/repos/:owner/:repo/downloads/:id"},
	{"DELETE", "/repos/:owner/:repo/downloads/:id"},
	{"GET", "/repos/:owner/:repo/forks"},
	{"POST", "/repos/:owner/:repo/forks"},
	{"GET", "/repos/:owner/:repo/hooks"},
	{"GET", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks"},
	//{"PATCH", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/hooks/:id/tests"},
	{"DELETE", "/repos/:owner/:repo/hooks/:id"},
	{"POST", "/repos/:owner/:repo/merges"},
	{"GET", "/repos/:owner/:repo/releases"},
	{"GET", "/repos/:owner/:repo/releases/:id"},
	{"POST", "/repos/:owner/:repo/releases"},
	//{"PATCH", "/repos/:owner/:repo/releases/:id"},
	{"DELETE", "/repos/:owner/:repo/releases/:id"},
	{"GET", "/repos/:owner/:repo/releases/:id/assets"},
	{"GET", "/repos/:owner/:repo/stats/contributors"},
	{"GET", "/repos/:owner/:repo/stats/commit_activity"},
	{"GET", "/repos/:owner/:repo/stats/code_frequency"},
	{"GET", "/repos/:owner/:repo/stats/participation"},
	{"GET", "/repos/:owner/:repo/stats/punch_card"},
	{"GET", "/repos/:owner/:repo/statuses/:ref"},
	{"POST", "/repos/:owner/:repo/statuses/:ref"},

	// Search
	{"GET", "/search/repositories"},
	{"GET", "/search/code"},
	{"GET", "/search/issues"},
	{"GET", "/search/users"},
	{"GET", "/legacy/issues/search/:owner/:repository/:state/:keyword"},
	{"GET", "/legacy/repos/search/:keyword"},
	{"GET", "/legacy/user/search/:keyword"},
	{"GET", "/legacy/user/email/:email"},

	// Users
	{"GET", "/users/:user"},
	{"GET", "/user"},
	//{"PATCH", "/user"},
	{"GET", "/users"},
	{"GET", "/user/emails"},
	{"POST", "/user/emails"},
	{"DELETE", "/user/emails"},
	{"GET", "/users/:user/followers"},
	{"GET", "/user/followers"},
	{"GET", "/users/:user/following"},
	{"GET", "/user/following"},
	{"GET", "/user/following/:user"},
	{"GET", "/users/:user/following/:target_user"},
	{"PUT", "/user/following/:user"},
	{"DELETE", "/user/following/:user"},
	{"GET", "/users/:user/keys"},
	{"GET", "/user/keys"},
	{"GET", "/user/keys/:id"},
	{"POST", "/user/keys"},
	//{"PATCH", "/user/keys/:id"},
	{"DELETE", "/user/keys/:id"},
}

var parseAPI = []testRoute{
	// Objects
	{"POST", "/1/classes/:className"},
	{"GET", "/1/classes/:className/:objectId"},
	{"PUT", "/1/classes/:className/:objectId"},
	{"GET", "/1/classes/:className"},
	{"DELETE", "/1/classes/:className/:objectId"},

	// Users
	{"POST", "/1/users"},
	{"GET", "/1/login"},
	{"GET", "/1/users/:objectId"},
	{"PUT", "/1/users/:objectId"},
	{"GET", "/1/users"},
	{"DELETE", "/1/users/:objectId"},
	{"POST", "/1/requestPasswordReset"},

	// Roles
	{"POST", "/1/roles"},
	{"GET", "/1/roles/:objectId"},
	{"PUT", "/1/roles/:objectId"},
	{"GET", "/1/roles"},
	{"DELETE", "/1/roles/:objectId"},

	// Files
	{"POST", "/1/files/:fileName"},

	// Analytics
	{"POST", "/1/events/:eventName"},

	// Push Notifications
	{"POST", "/1/push"},

	// Installations
	{"POST", "/1/installations"},
	{"GET", "/1/installations/:objectId"},
	{"PUT", "/1/installations/:objectId"},
	{"GET", "/1/installations"},
	{"DELETE", "/1/installations/:objectId"},

	// Cloud Functions
	{"POST", "/1/functions"},
}

var googlePlusAPI = []testRoute{
	// People
	{"GET", "/people/:userId"},
	{"GET", "/people"},
	{"GET", "/activities/:activityId/people/:collection"},
	{"GET", "/people/:userId/people/:collection"},
	{"GET", "/people/:userId/openIdConnect"},

	// Activities
	{"GET", "/people/:userId/activities/:collection"},
	{"GET", "/activities/:activityId"},
	{"GET", "/activities"},

	// Comments
	{"GET", "/activities/:activityId/comments"},
	{"GET", "/comments/:commentId"},

	// Moments
	{"POST", "/people/:userId/moments/:collection"},
	{"GET", "/people/:userId/moments/:collection"},
	{"DELETE", "/moments/:id"},
}

// handlerHelper 对应 echo 的 handlerHelper：handler 执行后写入 store 供断言。
// echo 用 c.Set/c.Path()，lazygo 用 c.WithValue/c.GetRoutePath()。
func handlerHelper(key string, value int) HandlerFunc {
	return func(c Context) error {
		c.WithValue(key, value)
		c.WithValue("path", c.GetRoutePath())
		return nil
	}
}

// newTestContext 创建带真实请求的测试 context。
// 注意：lazygo 的 c.Value() 在 store miss 时会回落到 request.Context()，
// request 为 nil 会 panic（与 echo 的 c.Get 仅查 store 不同），故统一传真实请求。
func newTestContext(s *Server) *context {
	return s.newContext(httptest.NewRequest(http.MethodGet, "/", nil), nil).(*context)
}

// findAndRun 查找路由并执行 handler（echo 测试里常见的 Find + c.handler(c) 组合）
func findAndRun(s *Server, method, path string, c *context) error {
	s.router.Find(method, path, c)
	return c.handler(c)
}

func TestRouterStatic(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	path := "/folders/a/files/echo.gif"
	r.Add(http.MethodGet, path, func(c Context) error {
		c.WithValue("path", path)
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, path, c)
	c.handler(c)
	assert.Equal(path, c.Value("path"))
}

func TestRouterParam(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/users/1", c)
	id, ok := c.Param("id")
	assert.True(ok)
	assert.Equal("1", id)
}

func TestRouterTwoParam(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(Context) error {
		return nil
	})
	c := newTestContext(s)

	r.Find(http.MethodGet, "/users/1/files/1", c)
	uid, ok := c.Param("uid")
	assert.True(ok)
	assert.Equal("1", uid)
	fid, ok := c.Param("fid")
	assert.True(ok)
	assert.Equal("1", fid)
}

// Issue #378
func TestRouterParamWithSlash(t *testing.T) {
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/a/:b/c/d/:e", func(c Context) error {
		return nil
	})

	r.Add(http.MethodGet, "/a/:b/c/:d/:f", func(c Context) error {
		return nil
	})

	c := newTestContext(s)
	assert.NotPanics(t, func() {
		r.Find(http.MethodGet, "/a/1/c/d/2/3", c)
	})
}

// Issue #1509
func TestRouterParamStaticConflict(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	handler := func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	}

	g := s.Group("/g")
	g.Get("/skills", handler)
	g.Get("/status", handler)
	g.Get("/:name", handler)

	c := newTestContext(s)
	r.Find(http.MethodGet, "/g/s", c)
	c.handler(c)
	name, ok := c.Param("name")
	assert.True(ok)
	assert.Equal("s", name)
	assert.Equal("/g/:name", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/g/status", c)
	c.handler(c)
	assert.Equal("/g/status", c.Value("path"))
}

func TestRouterMatchAny(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Routes
	r.Add(http.MethodGet, "/", func(Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/*", func(Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/*", func(Context) error {
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/", c)
	star, _ := c.Param("*")
	assert.Equal("", star)

	r.Find(http.MethodGet, "/download", c)
	star, _ = c.Param("*")
	assert.Equal("download", star)

	r.Find(http.MethodGet, "/users/joe", c)
	star, _ = c.Param("*")
	assert.Equal("joe", star)
}

// TestRouterMatchAnySlash shall verify finding the best route
// for any routes with trailing slash requests
func TestRouterMatchAnySlash(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	handler := func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	}

	// Routes
	r.Add(http.MethodGet, "/users", handler)
	r.Add(http.MethodGet, "/users/*", handler)
	r.Add(http.MethodGet, "/img/*", handler)
	r.Add(http.MethodGet, "/img/load", handler)
	r.Add(http.MethodGet, "/img/load/*", handler)
	r.Add(http.MethodGet, "/assets/*", handler)

	c := newTestContext(s)
	r.Find(http.MethodGet, "/", c)
	star, _ := c.Param("*")
	assert.Equal("", star)

	// Test trailing slash request for simple any route (see #1526)
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/", c)
	c.handler(c)
	assert.Equal("/users/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/joe", c)
	c.handler(c)
	assert.Equal("/users/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("joe", star)

	// Test trailing slash request for nested any route (see #1526)
	c = newTestContext(s)
	r.Find(http.MethodGet, "/img/load", c)
	c.handler(c)
	assert.Equal("/img/load", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/img/load/", c)
	c.handler(c)
	assert.Equal("/img/load/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/img/load/ben", c)
	c.handler(c)
	assert.Equal("/img/load/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("ben", star)

	// Test /assets/* any route
	// ... without trailing slash must not match
	c = newTestContext(s)
	r.Find(http.MethodGet, "/assets", c)
	c.handler(c)
	assert.Nil(c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)

	// ... with trailing slash must match
	c = newTestContext(s)
	r.Find(http.MethodGet, "/assets/", c)
	c.handler(c)
	assert.Equal("/assets/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)
}

func TestRouterMatchAnyMultiLevel(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	handler := func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	}

	// Routes
	r.Add(http.MethodGet, "/api/users/jack", handler)
	r.Add(http.MethodGet, "/api/users/jill", handler)
	r.Add(http.MethodGet, "/api/users/*", handler)
	r.Add(http.MethodGet, "/api/*", handler)
	r.Add(http.MethodGet, "/other/*", handler)
	r.Add(http.MethodGet, "/*", handler)

	c := newTestContext(s)
	r.Find(http.MethodGet, "/api/users/jack", c)
	c.handler(c)
	assert.Equal("/api/users/jack", c.Value("path"))
	star, _ := c.Param("*")
	assert.Equal("", star)

	r.Find(http.MethodGet, "/api/users/jill", c)
	c.handler(c)
	assert.Equal("/api/users/jill", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("", star)

	r.Find(http.MethodGet, "/api/users/joe", c)
	c.handler(c)
	assert.Equal("/api/users/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("joe", star)

	r.Find(http.MethodGet, "/api/nousers/joe", c)
	c.handler(c)
	assert.Equal("/api/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("nousers/joe", star)

	r.Find(http.MethodGet, "/api/none", c)
	c.handler(c)
	assert.Equal("/api/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("none", star)

	r.Find(http.MethodGet, "/noapi/users/jim", c)
	c.handler(c)
	assert.Equal("/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("noapi/users/jim", star)
}

func TestRouterMatchAnyMultiLevelWithPost(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	handler := func(c Context) error {
		c.WithValue("path", c.GetRoutePath())
		return nil
	}

	// Routes
	s.Post("/api/auth/login", handler)
	s.Post("/api/auth/forgotPassword", handler)
	s.Any("/api/*", handler)
	s.Any("/*", handler)

	// POST /api/auth/login shall choose login method
	c := newTestContext(s)
	r.Find(http.MethodPost, "/api/auth/login", c)
	c.handler(c)
	assert.Equal("/api/auth/login", c.Value("path"))
	star, _ := c.Param("*")
	assert.Equal("", star)

	// POST /api/auth/logout shall choose nearest any route
	c = newTestContext(s)
	r.Find(http.MethodPost, "/api/auth/logout", c)
	c.handler(c)
	assert.Equal("/api/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("auth/logout", star)

	// POST to /api/other/test shall choose nearest any route
	c = newTestContext(s)
	r.Find(http.MethodPost, "/api/other/test", c)
	c.handler(c)
	assert.Equal("/api/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("other/test", star)

	// GET to /api/other/test shall choose nearest any route
	c = newTestContext(s)
	r.Find(http.MethodGet, "/api/other/test", c)
	c.handler(c)
	assert.Equal("/api/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("other/test", star)
}

func TestRouterMicroParam(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	r.Add(http.MethodGet, "/:a/:b/:c", func(c Context) error {
		return nil
	})
	c := newTestContext(s)
	r.Find(http.MethodGet, "/1/2/3", c)
	a, _ := c.Param("a")
	b, _ := c.Param("b")
	cc, _ := c.Param("c")
	assert.Equal("1", a)
	assert.Equal("2", b)
	assert.Equal("3", cc)
}

func TestRouterMixParamMatchAny(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Route
	r.Add(http.MethodGet, "/users/:id/*", func(c Context) error {
		return nil
	})
	c := newTestContext(s)

	r.Find(http.MethodGet, "/users/joe/comments", c)
	c.handler(c)
	id, ok := c.Param("id")
	assert.True(ok)
	assert.Equal("joe", id)
}

func TestRouterMultiRoute(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Routes
	r.Add(http.MethodGet, "/users", func(c Context) error {
		c.WithValue("path", "/users")
		return nil
	})
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	c := newTestContext(s)

	// Route > /users
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal("/users", c.Value("path"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/1", c)
	id, _ := c.Param("id")
	assert.Equal("1", id)

	// Route > /user
	c = newTestContext(s)
	r.Find(http.MethodGet, "/user", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(http.StatusNotFound, he.Code)
}

func TestRouterPriority(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Routes
	r.Add(http.MethodGet, "/users", handlerHelper("a", 1))
	r.Add(http.MethodGet, "/users/new", handlerHelper("b", 2))
	r.Add(http.MethodGet, "/users/:id", handlerHelper("c", 3))
	r.Add(http.MethodGet, "/users/dew", handlerHelper("d", 4))
	r.Add(http.MethodGet, "/users/:id/files", handlerHelper("e", 5))
	r.Add(http.MethodGet, "/users/newsee", handlerHelper("f", 6))
	r.Add(http.MethodGet, "/users/*", handlerHelper("g", 7))
	r.Add(http.MethodGet, "/users/new/*", handlerHelper("h", 8))
	r.Add(http.MethodGet, "/*", handlerHelper("i", 9))

	// Route > /users
	c := newTestContext(s)
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal(1, c.Value("a"))
	assert.Equal("/users", c.Value("path"))

	// Route > /users/new
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/new", c)
	c.handler(c)
	assert.Equal(2, c.Value("b"))
	assert.Equal("/users/new", c.Value("path"))

	// Route > /users/:id
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/1", c)
	c.handler(c)
	assert.Equal(3, c.Value("c"))
	assert.Equal("/users/:id", c.Value("path"))

	// Route > /users/dew
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/dew", c)
	c.handler(c)
	assert.Equal(4, c.Value("d"))
	assert.Equal("/users/dew", c.Value("path"))

	// Route > /users/:id/files
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/1/files", c)
	c.handler(c)
	assert.Equal(5, c.Value("e"))
	assert.Equal("/users/:id/files", c.Value("path"))

	// Route > /users/:id
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/news", c)
	c.handler(c)
	assert.Equal(3, c.Value("c"))
	assert.Equal("/users/:id", c.Value("path"))

	// Route > /users/newsee
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/newsee", c)
	c.handler(c)
	assert.Equal(6, c.Value("f"))
	assert.Equal("/users/newsee", c.Value("path"))

	// Route > /users/newsee
	r.Find(http.MethodGet, "/users/newsee", c)
	c.handler(c)
	assert.Equal(6, c.Value("f"))

	// Route > /users/newsee
	r.Find(http.MethodGet, "/users/newsee", c)
	c.handler(c)
	assert.Equal(6, c.Value("f"))

	// Route > /users/*
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/joe/books", c)
	c.handler(c)
	assert.Equal(7, c.Value("g"))
	assert.Equal("/users/*", c.Value("path"))
	star, _ := c.Param("*")
	assert.Equal("joe/books", star)

	// Route > /users/new/* should be matched
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/new/someone", c)
	c.handler(c)
	assert.Equal(8, c.Value("h"))
	assert.Equal("/users/new/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("someone", star)

	// Route > /users/* should be matched although /users/dew exists
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/dew/someone", c)
	c.handler(c)
	assert.Equal(7, c.Value("g"))
	assert.Equal("/users/*", c.Value("path"))

	star, _ = c.Param("*")
	assert.Equal("dew/someone", star)

	// Route > /users/* should be matched although /users/dew exists
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/notexists/someone", c)
	c.handler(c)
	assert.Equal(7, c.Value("g"))
	assert.Equal("/users/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("notexists/someone", star)

	// Route > *
	c = newTestContext(s)
	r.Find(http.MethodGet, "/nousers", c)
	c.handler(c)
	assert.Equal(9, c.Value("i"))
	assert.Equal("/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("nousers", star)

	// Route > *
	c = newTestContext(s)
	r.Find(http.MethodGet, "/nousers/new", c)
	c.handler(c)
	assert.Equal(9, c.Value("i"))
	assert.Equal("/*", c.Value("path"))
	star, _ = c.Param("*")
	assert.Equal("nousers/new", star)
}

// Issue #1348
func TestRouterIssue1348(t *testing.T) {
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/:lang/", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:lang/dupa", func(c Context) error {
		return nil
	})
}

// Issue #372
func TestRouterPriorityNotFound(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router
	c := newTestContext(s)

	// Add
	r.Add(http.MethodGet, "/a/foo", func(c Context) error {
		c.WithValue("a", 1)
		return nil
	})
	r.Add(http.MethodGet, "/a/bar", func(c Context) error {
		c.WithValue("b", 2)
		return nil
	})

	// Find
	r.Find(http.MethodGet, "/a/foo", c)
	c.handler(c)
	assert.Equal(1, c.Value("a"))

	r.Find(http.MethodGet, "/a/bar", c)
	c.handler(c)
	assert.Equal(2, c.Value("b"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/abc/def", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(http.StatusNotFound, he.Code)
}

func TestRouterParamNames(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Routes
	r.Add(http.MethodGet, "/users", func(c Context) error {
		c.WithValue("path", "/users")
		return nil
	})
	r.Add(http.MethodGet, "/users/:id", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/:uid/files/:fid", func(c Context) error {
		return nil
	})
	c := newTestContext(s)

	// Route > /users
	r.Find(http.MethodGet, "/users", c)
	c.handler(c)
	assert.Equal("/users", c.Value("path"))

	// Route > /users/:id
	r.Find(http.MethodGet, "/users/1", c)
	assert.Equal("id", c.pnames[0])
	id, _ := c.Param("id")
	assert.Equal("1", id)

	// Route > /users/:uid/files/:fid
	r.Find(http.MethodGet, "/users/1/files/1", c)
	assert.Equal("uid", c.pnames[0])
	uid, _ := c.Param("uid")
	assert.Equal("1", uid)
	assert.Equal("fid", c.pnames[1])
	fid, _ := c.Param("fid")
	assert.Equal("1", fid)
}

// Issue #623 and #1406
func TestRouterStaticDynamicConflict(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodGet, "/dictionary/skills", handlerHelper("a", 1))
	r.Add(http.MethodGet, "/dictionary/:name", handlerHelper("b", 2))
	r.Add(http.MethodGet, "/users/new", handlerHelper("d", 4))
	r.Add(http.MethodGet, "/users/:name", handlerHelper("e", 5))
	r.Add(http.MethodGet, "/server", handlerHelper("c", 3))
	r.Add(http.MethodGet, "/", handlerHelper("f", 6))

	c := newTestContext(s)
	r.Find(http.MethodGet, "/dictionary/skills", c)
	c.Handler()(c)
	assert.Equal(1, c.Value("a"))
	assert.Equal("/dictionary/skills", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/dictionary/skillsnot", c)
	c.Handler()(c)
	assert.Equal(2, c.Value("b"))
	assert.Equal("/dictionary/:name", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/dictionary/type", c)
	c.Handler()(c)
	assert.Equal(2, c.Value("b"))
	assert.Equal("/dictionary/:name", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/server", c)
	c.Handler()(c)
	assert.Equal(3, c.Value("c"))
	assert.Equal("/server", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/new", c)
	c.Handler()(c)
	assert.Equal(4, c.Value("d"))
	assert.Equal("/users/new", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/new2", c)
	c.Handler()(c)
	assert.Equal(5, c.Value("e"))
	assert.Equal("/users/:name", c.Value("path"))

	c = newTestContext(s)
	r.Find(http.MethodGet, "/", c)
	c.Handler()(c)
	assert.Equal(6, c.Value("f"))
	assert.Equal("/", c.Value("path"))
}

// Issue #1348
func TestRouterParamBacktraceNotFound(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	// Add
	r.Add(http.MethodGet, "/:param1", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/foo", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/bar", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/:param1/bar/:param2", func(c Context) error {
		return nil
	})

	c := newTestContext(s)

	// Find
	r.Find(http.MethodGet, "/a", c)
	p1, _ := c.Param("param1")
	assert.Equal("a", p1)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/a/foo", c)
	p1, _ = c.Param("param1")
	assert.Equal("a", p1)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/a/bar", c)
	p1, _ = c.Param("param1")
	assert.Equal("a", p1)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/a/bar/b", c)
	p1, _ = c.Param("param1")
	assert.Equal("a", p1)
	p2, _ := c.Param("param2")
	assert.Equal("b", p2)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/a/bbbbb", c)
	he := c.handler(c).(*HTTPError)
	assert.Equal(http.StatusNotFound, he.Code)
}

func testRouterAPI(t *testing.T, api []testRoute) {
	s := New()
	r := s.router

	for _, route := range api {
		r.Add(route.Method, route.Path, func(c Context) error {
			return nil
		})
	}
	c := newTestContext(s)
	for _, route := range api {
		r.Find(route.Method, route.Path, c)
		tokens := strings.Split(route.Path[1:], "/")
		for _, token := range tokens {
			if token[0] == ':' {
				v, _ := c.Param(token[1:])
				assert.Equal(t, token, v)
			}
		}
	}
}

func TestRouterGitHubAPI(t *testing.T) {
	testRouterAPI(t, gitHubAPI)
}

// Issue #729
func TestRouterParamAlias(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/users/:userID/following"},
		{http.MethodGet, "/users/:userID/followedBy"},
		{http.MethodGet, "/users/:userID/follow"},
	}
	testRouterAPI(t, api)
}

// Issue #1052
func TestRouterParamOrdering(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/:a/:b/:c/:id"},
		{http.MethodGet, "/:a/:id"},
		{http.MethodGet, "/:a/:e/:id"},
	}
	testRouterAPI(t, api)
	api2 := []testRoute{
		{http.MethodGet, "/:a/:id"},
		{http.MethodGet, "/:a/:e/:id"},
		{http.MethodGet, "/:a/:b/:c/:id"},
	}
	testRouterAPI(t, api2)
	api3 := []testRoute{
		{http.MethodGet, "/:a/:b/:c/:id"},
		{http.MethodGet, "/:a/:e/:id"},
		{http.MethodGet, "/:a/:id"},
	}
	testRouterAPI(t, api3)
}

// Issue #1139
func TestRouterMixedParams(t *testing.T) {
	api := []testRoute{
		{http.MethodGet, "/teacher/:tid/room/suggestions"},
		{http.MethodGet, "/teacher/:id"},
	}
	testRouterAPI(t, api)
	api2 := []testRoute{
		{http.MethodGet, "/teacher/:id"},
		{http.MethodGet, "/teacher/:tid/room/suggestions"},
	}
	testRouterAPI(t, api2)
}

func TestRouterParseAPI(t *testing.T) {
	testRouterAPI(t, parseAPI)
}

func TestRouterGooglePlusAPI(t *testing.T) {
	testRouterAPI(t, googlePlusAPI)
}

// Issue #1466
func TestRouterParam1466(t *testing.T) {
	assert := assert.New(t)
	s := New()
	r := s.router

	r.Add(http.MethodPost, "/users/signup", func(c Context) error {
		return nil
	})
	r.Add(http.MethodPost, "/users/signup/bulk", func(c Context) error {
		return nil
	})
	r.Add(http.MethodPost, "/users/survey", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/:username", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/interests/:name/users", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/skills/:name/users", func(c Context) error {
		return nil
	})
	// Additional routes for Issue 1479
	r.Add(http.MethodGet, "/users/:username/likes/projects/ids", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/:username/profile", func(c Context) error {
		return nil
	})
	r.Add(http.MethodGet, "/users/:username/uploads/:type", func(c Context) error {
		return nil
	})

	c := newTestContext(s)

	r.Find(http.MethodGet, "/users/ajitem", c)
	v, _ := c.Param("username")
	assert.Equal("ajitem", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/sharewithme", c)
	v, _ = c.Param("username")
	assert.Equal("sharewithme", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/signup", c)
	v, _ = c.Param("username")
	assert.Equal("", v)

	// Additional assertions for #1479
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/sharewithme/likes/projects/ids", c)
	v, _ = c.Param("username")
	assert.Equal("sharewithme", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/ajitem/likes/projects/ids", c)
	v, _ = c.Param("username")
	assert.Equal("ajitem", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/sharewithme/profile", c)
	v, _ = c.Param("username")
	assert.Equal("sharewithme", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/ajitem/profile", c)
	v, _ = c.Param("username")
	assert.Equal("ajitem", v)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/sharewithme/uploads/self", c)
	v, _ = c.Param("username")
	assert.Equal("sharewithme", v)
	tp, _ := c.Param("type")
	assert.Equal("self", tp)

	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/ajitem/uploads/self", c)
	v, _ = c.Param("username")
	assert.Equal("ajitem", v)
	tp, _ = c.Param("type")
	assert.Equal("self", tp)

	// Issue #1493 - check for routing loop
	c = newTestContext(s)
	r.Find(http.MethodGet, "/users/tree/free", c)
	v, _ = c.Param("id")
	assert.Equal("", v)
	assert.Equal(0, c.responseWriter.Status)
}
