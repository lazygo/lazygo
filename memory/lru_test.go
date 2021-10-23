package memory

import (
	"strconv"
	"testing"
	"time"

	testify "github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {

	assert := testify.New(t)

	conf := []*Config{
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
	lru.Set("kkk", []byte("vvv"), time.Second)

	// Get
	val, ok := lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "vvv")
	// Set
	lru.Set("kkk", []byte("v2v2v2"), time.Second)
	val, ok = lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "v2v2v2")

	// Delete
	lru.Delete("kkk")
	val, ok = lru.Get("kkk")
	assert.False(ok)
	assert.Equal(val.String(), "")

	// Flush
	lru.Set("kkk", []byte("v2v2v2"), time.Second)
	assert.Equal(lru.Size(), uint64(6))
	lru.Flush()
	assert.Equal(lru.Size(), uint64(0))

	// 淘汰
	lru.Set("kkk", []byte("v2v2v2"), time.Second)
	lru.Set("xxx", []byte("98765432109876543210"), time.Second)
	val, ok = lru.Get("kkk")
	assert.True(ok)
	assert.Equal(val.String(), "v2v2v2")
	size := lru.Size()
	padding := []byte("01234567890123456789")
	l := uint64(len(padding))
	for i := 0; i < KB; i++ {
		lru.Set("kkk"+strconv.Itoa(i), padding, time.Second)
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
