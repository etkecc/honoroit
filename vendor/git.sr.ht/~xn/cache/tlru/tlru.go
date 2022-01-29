package tlru

import (
	"sync"
	"time"
)

// TLRU - Time aware Least Recently Used cache
type TLRU struct {
	sync.RWMutex
	max  int
	ttl  time.Duration
	data map[interface{}]*item
}

type item struct {
	v       interface{}
	expires int64
}

// New TLRU cache
func New(size int, ttl time.Duration) *TLRU {
	if size <= 0 {
		size = 1
	}
	tlru := &TLRU{
		max:  size,
		ttl:  ttl,
		data: make(map[interface{}]*item, size),
	}
	go tlru.cleanup()

	return tlru
}

func (c *TLRU) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	for range ticker.C {
		c.Lock()
		now := time.Now().UnixMicro() + c.ttl.Microseconds()
		for k, v := range c.data {
			if now >= v.expires {
				delete(c.data, k)
			}
		}
		c.Unlock()
	}
}

// removeLRU removes least recently used item
func (c *TLRU) removeLRU() {
	var key interface{}
	lru := time.Now().UnixMicro()
	for k, v := range c.data {
		if v.expires < lru {
			key = k
			lru = v.expires
		}
	}
	delete(c.data, key)
}

// Set an item to cache
func (c *TLRU) Set(key interface{}, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if len(c.data) == c.max {
		c.removeLRU()
	}
	c.data[key] = &item{v: value, expires: time.Now().UnixMicro()}
}

// Has check if an item exists in cache, without useness update
func (c *TLRU) Has(key interface{}) bool {
	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *TLRU) Get(key interface{}) interface{} {
	v, has := c.data[key]
	if !has {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	c.data[key].expires = time.Now().UnixMicro()

	return v.v
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
