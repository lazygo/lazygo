package memory

import (
	"strconv"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {

	assert := testify.New(t)

	conf := []Config{
		{
			Name:     "test",
			Capacity: 20 * KB,
		},
	}
	err := Init(conf)
	if err != nil {
		assert.Nil(err, err.Error())
	}

	lru, err := LRUCache("test")
	assert.Nil(err)

	// Set
	lru.Set("kkk", []byte("vvv"), 1)

	// Get
	val, ok := lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "vvv")
	// Set
	lru.Set("kkk", []byte("v2v2v2"), 1)
	val, ok = lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "v2v2v2")

	// Delete
	lru.Delete("kkk")
	val, ok = lru.Get("kkk")
	assert.False(ok)
	assert.Equal(val.String(), "")

	// Flush
	lru.Set("kkk", []byte("v2v2v2"), 1)
	assert.Equal(lru.Size(), uint64(6))
	lru.Flush()
	assert.Equal(lru.Size(), uint64(0))

	// 淘汰
	lru.Set("kkk", []byte("v2v2v2"), 1)
	lru.Set("xxx", []byte("98765432109876543210"), 1)
	val, ok = lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "v2v2v2")
	size := lru.Size()
	padding := []byte("01234567890123456789")
	l := uint64(len(padding))
	for i := 0; i < KB; i++ {
		lru.Set("kkk"+strconv.Itoa(i), padding, 1)
		size += l
		if size < lru.Capacity() {
			assert.Equal(lru.Size(), size)
		}
		if i == 50 {
			// 访问一次，不会被淘汰
			val, ok = lru.Get("xxx")
			assert.True(ok)
			assert.Equal(val.String(), "98765432109876543210")
		}
	}
	assert.Equal(lru.Size(), lru.Capacity())
	val, ok = lru.Get("kkk")
	assert.False(ok)
	assert.Equal(val.String(), "")
	val, ok = lru.Get("xxx")
	assert.True(ok)
	assert.Equal(val.String(), "98765432109876543210")

	// test timeout
	time.Sleep(2 * time.Second)
	val, ok = lru.Get("xxx")
	assert.False(ok)

}

// Set 已存在的 key 时应刷新过期时间
func TestLRUSetRefreshTTL(t *testing.T) {

	assert := testify.New(t)

	conf := []Config{
		{
			Name:     "test_refresh",
			Capacity: 20 * KB,
		},
	}
	err := Init(conf)
	assert.Nil(err)

	lru, err := LRUCache("test_refresh")
	assert.Nil(err)

	// 重复 Set 应延长 TTL
	lru.Set("refresh", []byte("v"), 1)
	time.Sleep(600 * time.Millisecond)
	lru.Set("refresh", []byte("v"), 1)
	time.Sleep(600 * time.Millisecond)
	val, ok := lru.Get("refresh")
	assert.True(ok)
	assert.Equal(val.String(), "v")

	// 对已过期但尚未懒删除的 key 重新 Set，应立即可读到
	lru.Set("expired", []byte("v"), 1)
	time.Sleep(1100 * time.Millisecond)
	err = lru.Set("expired", []byte("v2"), 1)
	assert.Nil(err)
	val, ok = lru.Get("expired")
	assert.True(ok)
	assert.Equal(val.String(), "v2")
}
