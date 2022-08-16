package tlru

import (
	"sync"
	"time"

	"git.sr.ht/~xn/cache/v2/utils"
)

// TLRU - Time aware Least Recently Used cache
type TLRU[V any] struct {
	sync.RWMutex
	max   int
	ttl   time.Duration
	data  map[string]*item[V]
	stale bool
}

type item[V any] struct {
	v       V
	used    int64
	expires int64
}

// New TLRU cache
func New[V any](size int, ttl time.Duration, stale bool) *TLRU[V] {
	if size <= 0 {
		size = 1
	}
	tlru := &TLRU[V]{
		max:   size,
		ttl:   ttl,
		stale: stale,
		data:  make(map[string]*item[V], size),
	}

	return tlru
}

// removeLRU removes least recently used item
func (c *TLRU[V]) removeLRU() {
	var key string
	lru := time.Now().UnixMicro()
	for k, v := range c.data {
		if v.used < lru {
			key = k
			lru = v.used
		}
	}
	delete(c.data, key)
}

// Set an item to cache
func (c *TLRU[V]) Set(key string, value V) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) >= c.max {
		c.removeLRU()
	}
	now := time.Now().UnixMicro()
	c.data[key] = &item[V]{v: value, expires: now + c.ttl.Microseconds(), used: now - 100}
}

// Has check if an item exists in cache, without useness update
func (c *TLRU[V]) Has(key string) bool {
	c.RLock()
	defer c.RUnlock()

	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *TLRU[V]) Get(key string) V {
	c.RLock()
	v, has := c.data[key]
	c.RUnlock()
	if !has {
		return utils.Zero[V]()
	}

	// normal way
	if time.Now().UnixMicro() < v.expires {
		c.Lock()
		defer c.Unlock()
		c.data[key].used = time.Now().UnixMicro()

		return v.v
	}

	// stale cache
	c.Lock()
	delete(c.data, key)
	c.Unlock()
	if c.stale {
		return v.v
	}
	return utils.Zero[V]()
}

// Remove an item from cache
func (c *TLRU[V]) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *TLRU[V]) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[string]*item[V], c.max)
}
