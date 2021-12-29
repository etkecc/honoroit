package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type loggerSuite struct {
	suite.Suite
}

func (s *loggerSuite) SetupTest() {
	s.T().Helper()
}

// catchStdout re-routes stdout to separate pipe before function and switches it back after that, returns text catched in separate pipe
func (s *loggerSuite) catchStdout(function func()) string {
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

func (s *loggerSuite) TestNew() {
	log := New("test", "example")

	s.True(log.level == INFO)
}

func (s *loggerSuite) TestGetLog() {
	logger := New("", "INFO")
	stdLogger := log.New(os.Stdout, "", 0)

	s.Equal(stdLogger, logger.GetLog())
}

func (s *loggerSuite) TestGetLevel() {
	text := "INFO"
	id := INFO
	logger := New("", text)

	s.Equal(id, logger.level)
	s.Equal(text, logger.GetLevel())
}

func (s *loggerSuite) TestFatal() {
	defer func() {
		if r := recover(); r == nil {
			s.Error(errors.New("the code did not panic"))
		}
	}()

	s.catchStdout(func() {
		logger := New("", "ERROR")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
		logger.Fatal("Test")
	})
}

func (s *loggerSuite) TestError() {
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

func (s *loggerSuite) TestError_Skip() {
	expected := ""

	actual := s.catchStdout(func() {
		logger := New("", "ERROR")
		logger.Error("recovery(): Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *loggerSuite) TestError_Level() {
	expected := ""

	actual := s.catchStdout(func() {
		logger := New("", "FATAL")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *loggerSuite) TestWarn() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "WARNING")
		logger.Error("Test")
		logger.Warnfln("Test")
		logger.Info("Test")
		logger.Debug("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *loggerSuite) TestInfo() {
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

func (s *loggerSuite) TestDebug() {
	expected := "ERROR Test\n"
	expected += "WARNING Test\n"
	expected += "INFO Test\n"
	expected += "DEBUG Test\n"

	actual := s.catchStdout(func() {
		logger := New("", "DEBUG")
		logger.Error("Test")
		logger.Warn("Test")
		logger.Info("Test")
		logger.Debugfln("Test")
		logger.Trace("Test")
	})

	s.Equal(expected, actual)
}

func (s *loggerSuite) TestTrace() {
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

func TestLogger(t *testing.T) {
	suite.Run(t, new(loggerSuite))
}
