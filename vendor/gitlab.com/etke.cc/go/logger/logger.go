package logger

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
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

var (
	txtLevelMap = map[string]int{
		"TRACE":   TRACE,
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"WARNING": WARNING,
		"ERROR":   ERROR,
		"FATAL":   FATAL,
	}
	levelMap = map[int]string{
		TRACE:   "TRACE",
		DEBUG:   "DEBUG",
		INFO:    "INFO",
		WARNING: "WARNING",
		ERROR:   "ERROR",
		FATAL:   "FATAL",
	}
	sentryLevelMap = map[int]sentry.Level{
		TRACE:   sentry.LevelDebug,
		DEBUG:   sentry.LevelDebug,
		INFO:    sentry.LevelInfo,
		WARNING: sentry.LevelWarning,
		ERROR:   sentry.LevelError,
		FATAL:   sentry.LevelFatal,
	}
)

// New creates new Logger object
func New(prefix string, level string) *Logger {
	levelID, ok := txtLevelMap[strings.ToUpper(level)]
	if !ok {
		levelID = INFO
	}

	return &Logger{log: log.New(os.Stdout, prefix, 0), level: levelID}
}

// GetLog returns underlying Logger object, useful in cases where log.Logger required
func (l *Logger) GetLog() *log.Logger {
	return l.log
}

// GetLevel (current)
func (l *Logger) GetLevel() string {
	return levelMap[l.level]
}

// Fatal log and exit
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log.Panicln("FATAL", fmt.Sprintf(message, args...))
}

// Error log
func (l *Logger) Error(message string, args ...interface{}) {
	// do not recover
	if strings.HasPrefix(message, "recovery()") {
		return
	}

	message = fmt.Sprintf(message, args...)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: l.log.Prefix(),
		Message:  message,
		Level:    sentryLevelMap[ERROR],
	})

	if l.level > ERROR {
		return
	}

	l.log.Println("ERROR", message)
}

// Warn log
func (l *Logger) Warn(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: l.log.Prefix(),
		Message:  message,
		Level:    sentryLevelMap[WARNING],
	})
	if l.level > WARNING {
		return
	}

	l.log.Println("WARNING", message)
}

// Warnfln for mautrix.Logger
func (l *Logger) Warnfln(message string, args ...interface{}) {
	l.Warn(message, args...)
}

// Info log
func (l *Logger) Info(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: l.log.Prefix(),
		Message:  message,
		Level:    sentryLevelMap[INFO],
	})
	if l.level > INFO {
		return
	}

	l.log.Println("INFO", message)
}

// Debug log
func (l *Logger) Debug(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: l.log.Prefix(),
		Message:  message,
		Level:    sentryLevelMap[DEBUG],
	})
	if l.level > DEBUG {
		return
	}

	l.log.Println("DEBUG", message)
}

// Debugfln for mautrix.Logger
func (l *Logger) Debugfln(message string, args ...interface{}) {
	l.Debug(message, args...)
}

// Trace log
func (l *Logger) Trace(message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: l.log.Prefix(),
		Message:  message,
		Level:    sentryLevelMap[TRACE],
	})
	if l.level > TRACE {
		return
	}

	l.log.Println("TRACE", message)
}
