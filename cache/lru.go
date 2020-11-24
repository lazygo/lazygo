package cache

import (
	"container/list"
	"sync"
	"time"
)

const KB = 1 << 10
const MB = 1 << 20

var LRUCache = NewLRUCache(10 * MB)

type LRU interface {
	Get(key string) (v Value, ok bool)
	Set(key string, value Value)
	Delete(key string) bool
	Clear()
}

// 定义Value类型为byte数组
type Value []byte

// 定义结构体item，此item即为上面图示中的Item的类型
type item struct {
	key           string
	value         Value
	size          int
	time_accessed time.Time
}

// 定义LRU缓存结构体
type LRUImpl struct {
	mu sync.Mutex
	// 此list为上面图示中的list，是一个双向链表
	list *list.List
	// 此table为上面图示中的table，其键为缓存的Key，值为Item在list中的地址
	table map[string]*list.Element
	// 缓存已使用大小
	size uint64
	// 缓存最大容量
	capacity uint64
}

// 检查容量并执行淘汰策略
func (lru *LRUImpl) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*item)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= uint64(delValue.size)
	}
}

//添加缓存
func (lru *LRUImpl) addNew(key string, value Value) {
	newItem := &item{key, value, len(value), time.Now()}
	element := lru.list.PushFront(newItem)
	lru.table[key] = element
	lru.size += uint64(newItem.size)
	lru.checkCapacity()
}

// 更新缓存
func (lru *LRUImpl) updateInplace(element *list.Element, value Value) {
	valueSize := len(value)
	sizeDiff := valueSize - element.Value.(*item).size
	element.Value.(*item).value = value
	element.Value.(*item).size = valueSize
	lru.size += uint64(sizeDiff)
	lru.list.MoveToFront(element)
	element.Value.(*item).time_accessed = time.Now()
	lru.checkCapacity()
}

// 获取缓存
func (lru *LRUImpl) Get(key string) (v Value, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}
	lru.list.MoveToFront(element)
	element.Value.(*item).time_accessed = time.Now()
	return element.Value.(*item).value, true
}

// 设置缓存
func (lru *LRUImpl) Set(key string, value Value) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value)
	}
}

// 删除缓存
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

// 清空缓存
func (lru *LRUImpl) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
}

// 构造函数
func NewLRUCache(capacity uint64) LRU {
	return &LRUImpl{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}
