package null

import "git.sr.ht/~xn/cache/v2/utils"

// Null is empty/stub cache client, usable for testing
type Null[V any] struct{}

// New creates new stub client
func New[V any]() *Null[V] {
	return &Null[V]{}
}

// Get does nothing
func (c *Null[V]) Get(_ string) V {
	return utils.Zero[V]()
}

// Set does nothing
func (c *Null[V]) Set(_ string, _ V) {}

// Has does nothing
func (c *Null[V]) Has(_ string) bool {
	return false
}

// Remove does nothing
func (c *Null[V]) Remove(_ string) {}

// Purge does nothing
func (c *Null[V]) Purge() {}
