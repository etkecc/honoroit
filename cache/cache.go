package cache

import (
	"sync"
	"time"
)

// Store in-memory
type Store struct {
	data sync.Map
	ttl  int64
	tick time.Duration
}

type entry struct {
	value   interface{}
	expires int64
}

// New in-memory cache
func New(ttl time.Duration) *Store {
	store := &Store{
		ttl:  int64(ttl / time.Second), // cahed item TTL in seconds
		tick: ttl / 2,                  // cleanup tick delay/frequency
	}
	go store.maintenance()

	return store
}

func (s *Store) maintenance() {
	ticker := time.NewTicker(s.tick)
	for range ticker.C {
		s.Cleanup()
	}
}

// Cleanup the data store
func (s *Store) Cleanup() {
	now := time.Now().UTC().Unix()
	s.data.Range(func(key, value interface{}) bool {
		item, ok := value.(*entry)
		// can't parse = don't need to save it
		if !ok {
			s.data.Delete(key)
			return true
		}
		if now >= item.expires {
			s.data.Delete(key)
		}
		return true
	})
}

// Set to cache
func (s *Store) Set(key string, value interface{}) {
	item := &entry{
		expires: time.Now().UTC().Unix() + s.ttl,
		value:   value,
	}
	s.data.Store(key, item)
}

// Get from cache
func (s *Store) Get(key string) interface{} {
	value, ok := s.data.Load(key)
	if !ok {
		return nil
	}

	item, ok := value.(*entry)
	if !ok {
		return nil
	}

	return item.value
}
