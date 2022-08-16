package lfu

import (
	"sync"
	"time"

	"git.sr.ht/~xn/cache/v2/utils"
)

// LFU - Least Frequently Used cache
type LFU[V any] struct {
	sync.RWMutex
	max  int
	data map[string]*item[V]
}

type item[V any] struct {
	v    V
	used int64
}

// New LFU cache
func New[V any](size int) *LFU[V] {
	if size <= 0 {
		size = 1
	}
	return &LFU[V]{
		max:  size,
		data: make(map[string]*item[V], size),
	}
}

// removeLFU removes least frequently used item
func (c *LFU[V]) removeLFU() {
	var key string
	lfu := time.Now().UnixMicro()
	for k, v := range c.data {
		if v.used < lfu {
			key = k
			lfu = v.used
		}
	}
	delete(c.data, key)
}

// Set an item to cache
func (c *LFU[V]) Set(key string, value V) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == c.max {
		c.removeLFU()
	}
	c.data[key] = &item[V]{v: value}
}

// Has check if an item exists in cache, without useness update
func (c *LFU[V]) Has(key string) bool {
	c.RLock()
	defer c.RUnlock()

	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *LFU[V]) Get(key string) V {
	c.RLock()
	v, has := c.data[key]
	c.RUnlock()
	if !has {
		return utils.Zero[V]()
	}

	c.Lock()
	defer c.Unlock()
	c.data[key].used++

	return v.v
}

// Remove an item from cache
func (c *LFU[V]) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *LFU[V]) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[string]*item[V], c.max)
}
