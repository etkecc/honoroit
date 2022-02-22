// Package memcached is a wrapper of github.com/bradfitz/gomemcache with Cache interface
package memcached

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
)

// Memcached cache
type Memcached struct {
	c   Client
	ttl int32
}

// New memcached client
// ttl is the cache expiration time, in seconds: either a relative
// time from now (up to 1 month), or an absolute Unix epoch time.
// Zero means the Item has no expiration time.
func New(ttl int32, servers ...string) *Memcached {
	return &Memcached{
		c:   newClient(servers...),
		ttl: ttl,
	}
}

// toString converts arbitrary interface{} to string
func (c *Memcached) toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// toBytes converts arbitrary interface{} to bytes
func (c *Memcached) toBytes(v interface{}) []byte {
	vBytes, ok := v.([]byte)
	if ok {
		return vBytes
	}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil
	}

	return buf.Bytes()
}

// Set item to cache
func (c *Memcached) Set(key interface{}, value interface{}) {
	item := &memcache.Item{
		Key:        c.toString(key),
		Value:      c.toBytes(value),
		Expiration: c.ttl,
	}

	c.c.Add(item) // nolint // interface
}

// Get item from cache
func (c *Memcached) Get(key interface{}) interface{} {
	item, err := c.c.Get(c.toString(key))
	if err != nil {
		return nil
	}

	if item != nil {
		return item.Value
	}

	return nil
}

// Has item in cache
func (c *Memcached) Has(key interface{}) bool {
	item, err := c.c.Get(c.toString(key))
	if err != nil {
		return false
	}
	if item == nil {
		return false
	}

	return true
}

// Remove item from cache
func (c *Memcached) Remove(key interface{}) {
	c.c.Delete(c.toString(key)) // nolint // interface
}

// Purge all cache
func (c *Memcached) Purge() {
	c.c.FlushAll() // nolint // interface
}
