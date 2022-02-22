package lfu

import (
	"sync"
	"time"
)

// LFU - Least Frequently Used cache
type LFU struct {
	sync.RWMutex
	max  int
	data map[interface{}]*item
}

type item struct {
	v    interface{}
	used int64
}

// New LFU cache
func New(size int) *LFU {
	if size <= 0 {
		size = 1
	}
	return &LFU{
		max:  size,
		data: make(map[interface{}]*item, size),
	}
}

// removeLFU removes least frequently used item
func (c *LFU) removeLFU() {
	var key interface{}
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
func (c *LFU) Set(key interface{}, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == c.max {
		c.removeLFU()
	}
	c.data[key] = &item{v: value}
}

// Has check if an item exists in cache, without useness update
func (c *LFU) Has(key interface{}) bool {
	c.RLock()
	defer c.RUnlock()

	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *LFU) Get(key interface{}) interface{} {
	c.RLock()
	v, has := c.data[key]
	c.RUnlock()
	if !has {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	c.data[key].used++

	return v.v
}

// Remove an item from cache
func (c *LFU) Remove(key interface{}) {
	c.Lock()
	defer c.Unlock()

	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *LFU) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[interface{}]*item, c.max)
}
