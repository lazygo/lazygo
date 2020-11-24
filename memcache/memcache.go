package memcache

import (
	"github.com/bradfitz/gomemcache/memcache"
)

type Memcache struct {
	name string
	mc   *memcache.Client
}

func NewMemcache(name string, mc *memcache.Client) *Memcache {
	return &Memcache{
		name: name,
		mc:   mc,
	}
}

// 获取memcached
func (m *Memcache) Conn() *memcache.Client {
	return m.mc
}

// 获取给定密钥的项，密钥的长度必须不超过250字节。
func (m *Memcache) Get(key string) []byte {
	var err error
	var item *memcache.Item
	item, err = m.mc.Get(key)
	if err == nil {
		return item.Value
	}
	if err == memcache.ErrCacheMiss {
		return nil
	}
	panic(err)
}

// GetMulti是Get的批处理版本
func (m *Memcache) GetMulti(keys []string) map[string][]byte {
	var err error
	var items map[string]*memcache.Item
	items, err = m.mc.GetMulti(keys)
	if err == nil {
		val := map[string][]byte{}
		for key, item := range items {
			val[key] = item.Value
		}
		return val
	}
	panic(err)
}

// 按增量键原子递增
// 返回值是递增或出错后的新值
// 如果该值在memcached中不存在，则错误为ErrCacheMiss
// memcached中的值必须是十进制数，否则将返回错误
func (m *Memcache) Increment(key string, delta uint64) uint64 {
	newValue, err := m.mc.Increment(key, delta)
	CheckError(err)
	return newValue
}

// 按增量原子递减键的值
// 返回值是递减或出错后的新值
// 如果该值在memcached中不存在，则错误为ErrCacheMiss
// memcached中的值必须是十进制数，否则将返回错误
func (m *Memcache) Decrement(key string, delta uint64) uint64 {
	newValue, err := m.mc.Decrement(key, delta)
	CheckError(err)
	return newValue
}

// 无条件地写入给定项
func (m *Memcache) Set(key string, value []byte, expiration int32) bool {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiration,
	}
	err := m.mc.Set(item)
	return err == nil
}

// 如果给定项的键值不存在，则写入该项。如果不满足该条件，则返回ErrNotStored
func (m *Memcache) Add(key string, value []byte, expiration int32) bool {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiration,
	}
	err := m.mc.Add(item)
	return err == nil
}

// 写入给定项，但仅当服务器*确实*已经保存此键的数据
func (m *Memcache) Replace(key string, value []byte, expiration int32) error {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiration,
	}
	return m.mc.Replace(item)
}

// 使用提供的键删除项。如果缓存中不存在该项，则返回错误ErrCacheMiss。
func (m *Memcache) Delete(key string) error {
	return m.mc.Delete(key)
}

// 更新给定密钥的有效期
// seconds参数是Unix时间戳，如果秒数小于1个月，则是该项将在未来过期的秒数。 0表示该项目没有过期时间
// 如果键不在缓存中，则返回ErrCacheMiss
// key的长度必须不超过250字节
func (m *Memcache) Touch(key string, seconds int32) error {
	return m.mc.Touch(key, seconds)
}
