package memory

import (
	"container/list"
	"sync"
	"time"
)

const KB = 1 << 10
const MB = 1 << 20

type LRU interface {
	Get(key string) (v Value, ok bool)
	Set(key string, val []byte, ttl time.Duration)
	Delete(key string) bool
	Flush()
	Size() uint64
	Capacity() uint64
}

// 定义Value类型为byte数组
type Value []byte

func (v Value) String() string {
	return string(v)
}

// 定义结构体item
type item struct {
	key          string
	value        Value
	size         int
	timeAccessed time.Time
	deadline     time.Time
}

// 定义LRU缓存结构体
type LRUImpl struct {
	name string
	mu   sync.Mutex
	// 双向链表
	list *list.List
	// 键为缓存的Key，值为Item在list中的地址
	table map[string]*list.Element
	// 缓存已使用大小
	size uint64
	// 缓存最大容量
	capacity uint64
}

// checkCapacity 检查容量并执行淘汰策略
func (lru *LRUImpl) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*item)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= uint64(delValue.size)
	}
}

// addNew 添加缓存
func (lru *LRUImpl) addNew(key string, value Value, ttl time.Duration) {
	newItem := &item{
		key:          key,
		value:        value,
		size:         len(value),
		timeAccessed: time.Now(),
		deadline:     time.Now().Add(ttl),
	}
	element := lru.list.PushFront(newItem)
	lru.table[key] = element
	lru.size += uint64(newItem.size)
	lru.checkCapacity()
}

// updateInplace 更新缓存
func (lru *LRUImpl) updateInplace(element *list.Element, value Value) {
	valueSize := len(value)
	sizeDiff := valueSize - element.Value.(*item).size
	element.Value.(*item).value = value
	element.Value.(*item).size = valueSize
	lru.size += uint64(sizeDiff)
	lru.list.MoveToFront(element)
	element.Value.(*item).timeAccessed = time.Now()
	lru.checkCapacity()
}

// Get 获取缓存
func (lru *LRUImpl) Get(key string) (v Value, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}
	// 判断过期
	if element.Value.(*item).deadline.Before(time.Now()) {
		lru.list.Remove(element)
		delete(lru.table, key)
		lru.size -= uint64(element.Value.(*item).size)
		return nil, false
	}

	lru.list.MoveToFront(element)
	element.Value.(*item).timeAccessed = time.Now()
	return element.Value.(*item).value, true
}

// Set 设置缓存
func (lru *LRUImpl) Set(key string, val []byte, ttl time.Duration) {
	value := Value(val)

	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value, ttl)
	}
}

// Delete 删除缓存
func (lru *LRUImpl) Delete(key string) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= uint64(element.Value.(*item).size)
	return true
}

// Flush 清空缓存
func (lru *LRUImpl) Flush() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
}

// Size 获取lru已使用大小
func (lru *LRUImpl) Size() uint64 {
	return lru.size
}

// Capacity 获取lru容量
func (lru *LRUImpl) Capacity() uint64 {
	return lru.capacity
}

// newLRUCache 构造函数
func newLRUCache(name string, capacity uint64) LRU {
	return &LRUImpl{
		name:     name,
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}
