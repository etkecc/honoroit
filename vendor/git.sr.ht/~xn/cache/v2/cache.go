package cache

import (
	"time"

	"git.sr.ht/~xn/cache/v2/lfu"
	"git.sr.ht/~xn/cache/v2/lru"
	"git.sr.ht/~xn/cache/v2/null"
	"git.sr.ht/~xn/cache/v2/tlru"
)

// Cache interface, used by any cache implementation
type Cache[V any] interface {
	// Get an item from cache
	Get(key string) V
	// Set an item to cache
	Set(key string, value V)
	// Has an item in cache
	Has(key string) bool
	// Remove an item from cache
	Remove(key string)
	// Purge all items from cache
	Purge()
}

// NewNull creates an empty cache client, usable for testing
func NewNull[V any]() Cache[V] {
	return null.New[V]()
}

// NewLRU creates new Least Recently Used cache
func NewLRU[V any](size int) Cache[V] {
	return lru.New[V](size)
}

// NewTLRU cerates new Time aware Least Recently Used cache
// Arguments:
// size - max amount if items in cache
// ttl - cached item expiration duration
// stale - if true, stale cached item will be returned instead of nil, but only once
func NewTLRU[V any](size int, ttl time.Duration, stale bool) Cache[V] {
	return tlru.New[V](size, ttl, stale)
}

// NewLFU creates new Least Frequently Used cache
func NewLFU[V any](size int) Cache[V] {
	return lfu.New[V](size)
}
