package lru

import (
	"sync"
	"time"
)

// LRU - Least Recently Used cache
type LRU struct {
	sync.RWMutex
	max  int
	data map[interface{}]*item
}

type item struct {
	v    interface{}
	used int64
}

// New LRU cache
func New(size int) *LRU {
	if size <= 0 {
		size = 1
	}
	return &LRU{
		max:  size,
		data: make(map[interface{}]*item, size),
	}
}

// removeLRU removes least recently used item
func (c *LRU) removeLRU() {
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
func (c *LRU) Set(key interface{}, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if len(c.data) == c.max {
		c.removeLRU()
	}
	c.data[key] = &item{v: value, used: time.Now().UnixMicro()}
}

// Has check if an item exists in cache, without useness update
func (c *LRU) Has(key interface{}) bool {
	_, has := c.data[key]
	return has
}

// Get an item from cache
func (c *LRU) Get(key interface{}) interface{} {
	v, has := c.data[key]
	if !has {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	c.data[key].used = time.Now().UnixMicro()

	return v.v
}

// Remove an item from cache
func (c *LRU) Remove(key interface{}) {
	c.Lock()
	defer c.Unlock()
	if len(c.data) == 0 {
		return
	}
	delete(c.data, key)
}

// Purge cache
func (c *LRU) Purge() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[interface{}]*item, c.max)
}
