package memcached

import "github.com/bradfitz/gomemcache/memcache"

// Client for memcached protocol
type Client interface {
	Add(*memcache.Item) error
	Get(key string) (*memcache.Item, error)
	Delete(string) error
	FlushAll() error
}

// null client
type null struct{}

func newClient(servers ...string) Client {
	var client Client
	client = &null{}
	if len(servers) > 0 && servers[0] != "" {
		client = memcache.New(servers...)
	}

	return client
}

// Add does nothing
func (c *null) Add(_ *memcache.Item) error {
	return nil
}

// Get does nothing
func (c *null) Get(_ string) (*memcache.Item, error) {
	return nil, nil
}

// Delete does nothing
func (c *null) Delete(_ string) error {
	return nil
}

// FlushAll does nothing
func (c *null) FlushAll() error {
	return nil
}
