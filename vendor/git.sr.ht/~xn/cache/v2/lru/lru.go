package lru

import (
	"sync"
	"time"

	"git.sr.ht/~xn/cache/v2/utils"
)

// LRU - Least Recently Used cache
type LRU[V any] struct {
	sync.RWMutex
	max  int
	data map[string]*item[V]
}

type item[V any] struct {
	v    V
	used int64
}

// New LRU cache
func New[V any](size int) *LRU[V] {
	if size <= 0 {
		size = 1
	}
	return &LRU[V]{
		max:  size,
		data: make(map[string]*item[V], size),
	}
}

// removeLRU removes least recently used item
func (c *LRU[V]) removeLRU() {
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
func (c *LRU[V]) Set(key string, value V) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == c.max {
		c.removeLRU()
	}
	c.data[key] = &item[V]{v: value, used: time.Now().UnixMicro()}
}

// Has check if an item exists in cache, without useness update
func (c *LRU[V]) Has(key string) bool {
	c.RLock()
	defer c.RUnlock()

	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *LRU[V]) Get(key string) V {
	c.RLock()
	v, has := c.data[key]
	c.RUnlock()
	if !has {
		return utils.Zero[V]()
	}

	c.Lock()
	defer c.Unlock()
	c.data[key].used = time.Now().UnixMicro()

	return v.v
}

// Remove an item from cache
func (c *LRU[V]) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *LRU[V]) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[string]*item[V], c.max)
}
