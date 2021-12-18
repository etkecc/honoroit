package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Logger struct
type Logger struct {
	log   *log.Logger
	level int
}

const (
	// TRACE level
	TRACE int = iota
	// DEBUG level
	DEBUG
	// INFO level
	INFO
	// WARNING level
	WARNING
	// ERROR level
	ERROR
	// FATAL level
	FATAL
)

var levelMap = map[string]int{
	"TRACE":   TRACE,
	"DEBUG":   DEBUG,
	"INFO":    INFO,
	"WARNING": WARNING,
	"ERROR":   ERROR,
	"FATAL":   FATAL,
}

// New creates new Logger object
func New(prefix string, level string) *Logger {
	levelID, ok := levelMap[strings.ToUpper(level)]
	if !ok {
		levelID = INFO
	}

	return &Logger{log: log.New(os.Stdout, prefix, 0), level: levelID}
}

// GetLog returns underlying Logger object, useful in cases where log.Logger required
func (l *Logger) GetLog() *log.Logger {
	return l.log
}

// Fatal log and exit
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log.Panicln("FATAL", fmt.Sprintf(message, args...))
}

// Error log
func (l *Logger) Error(message string, args ...interface{}) {
	if l.level > ERROR {
		return
	}

	l.log.Println("ERROR", fmt.Sprintf(message, args...))
}

// Warn log
func (l *Logger) Warn(message string, args ...interface{}) {
	if l.level > WARNING {
		return
	}

	l.log.Println("WARNING", fmt.Sprintf(message, args...))
}

// Info log
func (l *Logger) Info(message string, args ...interface{}) {
	if l.level > INFO {
		return
	}

	l.log.Println("INFO", fmt.Sprintf(message, args...))
}

// Debug log
func (l *Logger) Debug(message string, args ...interface{}) {
	if l.level > DEBUG {
		return
	}

	l.log.Println("DEBUG", fmt.Sprintf(message, args...))
}

// Trace log
func (l *Logger) Trace(message string, args ...interface{}) {
	if l.level > TRACE {
		return
	}

	l.log.Println("TRACE", fmt.Sprintf(message, args...))
}
