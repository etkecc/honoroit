package tlru

import (
	"sync"
	"time"
)

// TLRU - Time aware Least Recently Used cache
type TLRU struct {
	sync.RWMutex
	max   int
	ttl   time.Duration
	data  map[interface{}]*item
	stale bool
}

type item struct {
	v       interface{}
	used    int64
	expires int64
}

// New TLRU cache
func New(size int, ttl time.Duration, stale bool) *TLRU {
	if size <= 0 {
		size = 1
	}
	tlru := &TLRU{
		max:   size,
		ttl:   ttl,
		stale: stale,
		data:  make(map[interface{}]*item, size),
	}

	return tlru
}

// removeLRU removes least recently used item
func (c *TLRU) removeLRU() {
	var key interface{}
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
func (c *TLRU) Set(key interface{}, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) >= c.max {
		c.removeLRU()
	}
	now := time.Now().UnixMicro()
	c.data[key] = &item{v: value, expires: now + c.ttl.Microseconds(), used: now - 100}
}

// Has check if an item exists in cache, without useness update
func (c *TLRU) Has(key interface{}) bool {
	c.RLock()
	defer c.RUnlock()

	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *TLRU) Get(key interface{}) interface{} {
	c.RLock()
	v, has := c.data[key]
	c.RUnlock()
	if !has {
		return nil
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
	return nil
}

// Remove an item from cache
func (c *TLRU) Remove(key interface{}) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *TLRU) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[interface{}]*item, c.max)
}
