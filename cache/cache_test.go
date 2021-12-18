package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type cacheSuite struct {
	suite.Suite
	store *Store
}

func (s *cacheSuite) SetupTest() {
	s.T().Helper()
	s.store = New(1 * time.Second)
}

func (s *cacheSuite) TestGet() {
	key := "key"
	value := "value"
	item := &entry{
		expires: time.Now().Unix() + 1,
		value:   value,
	}
	s.store.data.Store(key, item)

	valueInterface := s.store.Get(key)
	actual, ok := valueInterface.(string)

	s.True(ok)
	s.Equal(value, actual)
}

func (s *cacheSuite) TestGet_Invalid() {
	key := "key"
	value := "value"
	s.store.data.Store(key, value)

	valueInterface := s.store.Get(key)
	actual, ok := valueInterface.(string)

	s.Nil(valueInterface)
	s.False(ok)
	s.Empty(actual)
}

func (s *cacheSuite) TestSet() {
	key := "key"
	value := "value"
	item := &entry{
		expires: time.Now().Unix() + 1,
		value:   value,
	}

	s.store.Set(key, value)

	itemInterface, loadOK := s.store.data.Load(key)
	actual, castOK := itemInterface.(*entry)

	s.True(loadOK)
	s.True(castOK)
	s.Equal(item, actual)
}

func (s *cacheSuite) TestCleanup() {
	key := "key"
	item := &entry{
		expires: time.Now().Unix() - 1,
		value:   "value",
	}
	s.store.data.Store(key, item)

	s.store.Cleanup()

	valueInterface := s.store.Get(key)
	actual, ok := valueInterface.(string)
	s.Nil(valueInterface)
	s.False(ok)
	s.Empty(actual)
}

func (s *cacheSuite) TestCleanup_Invalid() {
	key := "key"
	item := "value"
	s.store.data.Store(key, item)

	s.store.Cleanup()

	valueInterface := s.store.Get(key)
	actual, ok := valueInterface.(string)
	s.Nil(valueInterface)
	s.False(ok)
	s.Empty(actual)
}

func TestCache(t *testing.T) {
	suite.Run(t, new(cacheSuite))
}
