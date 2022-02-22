package cache

import (
	"time"

	"git.sr.ht/~xn/cache/lfu"
	"git.sr.ht/~xn/cache/lru"
	"git.sr.ht/~xn/cache/memcached"
	"git.sr.ht/~xn/cache/null"
	"git.sr.ht/~xn/cache/tlru"
)

// Cache interface, used by any cache implementation
type Cache interface {
	// Get an item from cache
	Get(key interface{}) interface{}
	// Set an item to cache
	Set(key interface{}, value interface{})
	// Has an item in cache
	Has(key interface{}) bool
	// Remove an item from cache
	Remove(key interface{})
	// Purge all items from cache
	Purge()
}

// NewNull creates an empty cache client, usable for testing
func NewNull() Cache {
	return null.New()
}

// NewLRU creates new Least Recently Used cache
func NewLRU(size int) Cache {
	return lru.New(size)
}

// NewTLRU cerates new Time aware Least Recently Used cache
func NewTLRU(size int, ttl time.Duration) Cache {
	return tlru.New(size, ttl)
}

// NewLFU creates new Least Frequently Used cache
func NewLFU(size int) Cache {
	return lfu.New(size)
}

// NewMemcached creates new Memcached client
func NewMemcached(ttl int32, servers ...string) Cache {
	return memcached.New(ttl, servers...)
}
