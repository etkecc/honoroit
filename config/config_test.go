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
	"HONOROIT_TOKEN":      "test",
	"HONOROIT_ROOMID":     "!test:example.com",

	"HONOROIT_LOGLEVEL": "DEBUG",

	"HONOROIT_TEXT_GREETINGS": "hello",
	"HONOROIT_TEXT_ERROR":     "error",
	"HONOROIT_TEXT_EMPTYROOM": "empty room",
	"HONOROIT_TEXT_DONE":      "done",
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
	s.Equal("test", config.Token)
	s.Equal("!test:example.com", config.RoomID)
	s.Equal("DEBUG", config.LogLevel)
	s.Equal("hello", config.Text.Greetings)
	s.Equal("error", config.Text.Error)
	s.Equal("empty room", config.Text.EmptyRoom)
	s.Equal("done", config.Text.Done)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(configSuite))
}
