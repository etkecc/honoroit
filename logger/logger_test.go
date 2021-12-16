package logger

import (
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggerSuite struct {
	suite.Suite
}

func (s *LoggerSuite) SetupTest() {
	s.T().Helper()
}

// catchStdout re-routes stdout to separate pipe before function and switches it back after that, returns text catched in separate pipe
func (s *LoggerSuite) catchStdout(function func()) string {
	s.T().Helper()

	osStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	function()

	w.Close()
	out, err := io.ReadAll(r)
	s.NoError(err)
	os.Stdout = osStdout

	return string(out)
}

func (s *LoggerSuite) TestNew() {
	log := New("test", "example")

	s.True(log.level == INFO)
}

func (s *LoggerSuite) TestGetLog() {
	logger := New("", "INFO")
	stdLogger := log.New(os.Stdout, "", 0)

	s.Equal(stdLogger, logger.GetLog())
}

func (s *LoggerSuite) TestError() {
	expected := "ERROR Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "ERROR")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *LoggerSuite) TestWarn() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "WARNING")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *LoggerSuite) TestInfo() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"
	expected += "INFO Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "INFO")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *LoggerSuite) TestDebug() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"
	expected += "INFO Test\n"
	expected += "DEBUG Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "DEBUG")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *LoggerSuite) TestTrace() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"
	expected += "INFO Test\n"
	expected += "DEBUG Test\n"
	expected += "TRACE Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "TRACE")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
