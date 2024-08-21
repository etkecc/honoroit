package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type configSuite struct {
	suite.Suite
}

var values = map[string]string{
	"HONOROIT_HOMESERVER": "https://example.com",
	"HONOROIT_LOGIN":      "test",
	"HONOROIT_PASSWORD":   "password",
	"HONOROIT_ROOMID":     "!test:example.com",

	"HONOROIT_PREFIX":    "!hohoho",
	"HONOROIT_LOGLEVEL":  "DEBUG",
	"HONOROIT_CACHESIZE": "100",

	"HONOROIT_DB_DIALECT": "sqlite3",
	"HONOROIT_DB_DSN":     "/tmp/test.db",
}

func (s *configSuite) SetupTest() {
	s.T().Helper()
	for key, value := range values {
		os.Setenv(key, value)
	}
}

func (s *configSuite) TearDownTest() {
	s.T().Helper()
	for key := range values {
		os.Unsetenv(key)
	}
}

func (s *configSuite) TestNew() {
	config := New()

	s.Equal("https://example.com", config.Homeserver)
	s.Equal("test", config.Login)
	s.Equal("password", config.Password)
	s.Equal("!test:example.com", config.RoomID)
	s.Equal("DEBUG", config.LogLevel)
	s.Equal(100, config.CacheSize)
	s.Equal("!hohoho", config.Prefix)
	s.Equal("sqlite3", config.DB.Dialect)
	s.Equal("/tmp/test.db", config.DB.DSN)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(configSuite))
}
