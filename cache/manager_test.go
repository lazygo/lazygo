package cache

// 泛型包装与lru适配器的测试（无需外部服务）

import (
	"errors"
	"testing"
	"time"

	"github.com/lazygo/lazygo/memory"
	testify "github.com/stretchr/testify/assert"
)

type cacheUser struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func setupCache(t *testing.T) {
	assert := testify.New(t)
	err := memory.Init([]memory.Config{
		{Name: "cache_test", Capacity: 100},
	})
	assert.Nil(err)
	err = Init([]Config{
		{
			Name:    "cache_test",
			Adapter: "memory",
			Option:  map[string]string{"name": "cache_test", "prefix": "t_"},
		},
	}, "cache_test")
	assert.Nil(err)
}

func TestGenericGet(t *testing.T) {
	assert := testify.New(t)
	setupCache(t)

	// 未命中
	_, ok, err := Get[cacheUser]("user_1")
	assert.Nil(err)
	assert.False(ok)

	// Set后命中
	err = Set("user_1", cacheUser{Id: 1, Name: "a"}, 60)
	assert.Nil(err)
	user, ok, err := Get[cacheUser]("user_1")
	assert.Nil(err)
	assert.True(ok)
	assert.Equal(uint64(1), user.Id)
	assert.Equal("a", user.Name)

	// 切片类型
	err = Set("user_list", []cacheUser{{Id: 2, Name: "b"}}, 60)
	assert.Nil(err)
	list, ok, err := Get[[]cacheUser]("user_list")
	assert.Nil(err)
	assert.True(ok)
	assert.Equal(1, len(list))
	assert.Equal("b", list[0].Name)
}

func TestGenericRemember(t *testing.T) {
	assert := testify.New(t)
	setupCache(t)

	calls := 0
	// 第一次：未命中，调用fn并写入
	user, hit, err := Remember[cacheUser]("remember_1", func() (cacheUser, error) {
		calls++
		return cacheUser{Id: 10, Name: "x"}, nil
	}, 60)
	assert.Nil(err)
	assert.False(hit)
	assert.Equal(uint64(10), user.Id)
	assert.Equal(1, calls)

	// 第二次：命中缓存，fn不再调用
	user, hit, err = Remember[cacheUser]("remember_1", func() (cacheUser, error) {
		calls++
		return cacheUser{Id: 10, Name: "x"}, nil
	}, 60)
	assert.Nil(err)
	assert.True(hit)
	assert.Equal("x", user.Name)
	assert.Equal(1, calls)

	// fn返回错误时透传
	sentinel := errors.New("boom")
	_, _, err = Remember[cacheUser]("remember_2", func() (cacheUser, error) {
		return cacheUser{}, sentinel
	}, 60)
	assert.ErrorIs(err, sentinel)
}

func TestGenericTTL(t *testing.T) {
	assert := testify.New(t)
	setupCache(t)

	err := Set("ttl_key", "v", 1)
	assert.Nil(err)
	_, ok, err := Get[string]("ttl_key")
	assert.Nil(err)
	assert.True(ok)

	// lru适配器的过期检查是惰性的，触发一次过期清理
	time.Sleep(1100 * time.Millisecond)
	_, ok, err = Get[string]("ttl_key")
	assert.Nil(err)
	assert.False(ok)
}

// 泛型Instance的强类型往返
func TestTypedInstance(t *testing.T) {
	assert := testify.New(t)
	setupCache(t)

	c, err := Instance[cacheUser]("cache_test")
	assert.Nil(err)

	// Get 未命中
	_, ok, err := c.Get("typed_1")
	assert.Nil(err)
	assert.False(ok)

	// Set 后强类型往返
	err = c.Set("typed_1", cacheUser{Id: 7, Name: "typed"}, 60)
	assert.Nil(err)
	user, ok, err := c.Get("typed_1")
	assert.Nil(err)
	assert.True(ok)
	assert.Equal(uint64(7), user.Id)
	assert.Equal("typed", user.Name)

	// Remember 未命中：调用fn并写入
	calls := 0
	user, hit, err := c.Remember("typed_2", func() (cacheUser, error) {
		calls++
		return cacheUser{Id: 8, Name: "r"}, nil
	}, 60)
	assert.Nil(err)
	assert.False(hit)
	assert.Equal(uint64(8), user.Id)
	assert.Equal(1, calls)

	// Remember 命中：fn不再调用
	user, hit, err = c.Remember("typed_2", func() (cacheUser, error) {
		calls++
		return cacheUser{}, nil
	}, 60)
	assert.Nil(err)
	assert.True(hit)
	assert.Equal("r", user.Name)
	assert.Equal(1, calls)

	// Has / HasMulti / Forget 直接委托backend
	ok, err = c.Has("typed_1")
	assert.Nil(err)
	assert.True(ok)
	m, err := c.HasMulti("typed_1", "typed_none")
	assert.Nil(err)
	assert.True(m["typed_1"])
	assert.False(m["typed_none"])
	err = c.Forget("typed_1")
	assert.Nil(err)
	_, ok, err = c.Get("typed_1")
	assert.Nil(err)
	assert.False(ok)
}
