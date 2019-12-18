package memcache

import (
	"github.com/bradfitz/gomemcache/memcache"
)

var LostConnection = []string{
	"reset by peer",
	"connection refused",
}

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

func (m *Memcache) Conn() *memcache.Client {
	return m.mc
}

// Get gets the item for the given key.
// The key must be at most 250 bytes in length.
func (m *Memcache) Get(key string) string {

	var err error
	var item *memcache.Item
	for retry := 2; retry > 0; retry-- {
		item, err = m.mc.Get(key)
		if err == nil {
			return string(item.Value)
		}
		if err == memcache.ErrCacheMiss {
			return ""
		}
		if !ContainInArray(err.Error(), LostConnection) {
			break
		}
	}
	panic(err)
}

// GetMulti is a batch version of Get. The returned map from keys to
// items may have fewer elements than the input slice, due to memcache
// cache misses. Each key must be at most 250 bytes in length.
// If no error is returned, the returned map will also be non-nil.
func (m *Memcache) GetMulti(keys []string) map[string]string {

	var err error
	var items map[string]*memcache.Item
	for retry := 2; retry > 0; retry-- {
		items, err = m.mc.GetMulti(keys)
		if err == nil {
			val := map[string]string{}
			for key, item := range items {
				val[key] = string(item.Value)
			}
			return val
		}
		if !ContainInArray(err.Error(), LostConnection) {
			break
		}
	}
	panic(err)
}

// Increment atomically increments key by delta. The return value is
// the new value after being incremented or an error. If the value
// didn't exist in memcached the error is ErrCacheMiss. The value in
// memcached must be an decimal number, or an error will be returned.
// On 64-bit overflow, the new value wraps around.
func (m *Memcache) Increment(key string, delta uint64) uint64 {
	newValue, err := m.mc.Increment(key, delta)
	CheckError(err)
	return newValue
}

// Decrement atomically decrements key by delta. The return value is
// the new value after being decremented or an error. If the value
// didn't exist in memcached the error is ErrCacheMiss. The value in
// memcached must be an decimal number, or an error will be returned.
// On underflow, the new value is capped at zero and does not wrap
// around.
func (m *Memcache) Decrement(key string, delta uint64) uint64 {
	newValue, err := m.mc.Decrement(key, delta)
	CheckError(err)
	return newValue
}

// Set writes the given item, unconditionally.
func (m *Memcache) Set(key string, value string, expiration int32) bool {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: expiration,
	}
	err := m.mc.Set(item)
	return err == nil
}

// Add writes the given item, if no value already exists for its
// key. ErrNotStored is returned if that condition is not met.
func (m *Memcache) Add(key string, value string, expiration int32) bool {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: expiration,
	}
	err := m.mc.Add(item)
	return err == nil
}

// Replace writes the given item, but only if the server *does*
// already hold data for this key
func (m *Memcache) Replace(key string, value string, expiration int32) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: expiration,
	}
	return m.mc.Replace(item)
}

// Delete deletes the item with the provided key. The error ErrCacheMiss is
// returned if the item didn't already exist in the cache.
func (m *Memcache) Delete(key string) error {
	return m.mc.Delete(key)
}

// Touch updates the expiry for the given key. The seconds parameter is either
// a Unix timestamp or, if seconds is less than 1 month, the number of seconds
// into the future at which time the item will expire. Zero means the item has
// no expiration time. ErrCacheMiss is returned if the key is not in the cache.
// The key must be at most 250 bytes in length.
func (m *Memcache) Touch(key string, seconds int32) error {
	return m.mc.Touch(key, seconds)
}
