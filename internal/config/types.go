package config

import (
	"time"

	echobasicauth "github.com/etkecc/go-echo-basic-auth"
)

// Config of Honoroit
type Config struct {
	// Homeserver url
	Homeserver string
	// Login is a MXID localpart (honoroit - OK, @honoroit:example.com - wrong)
	Login string
	// Password for login/password auth only
	Password string
	// SharedSecret for shared secret auth only
	SharedSecret string
	// DataSecret for account data encryption
	DataSecret string
	// NoEncryptionWarning disables warning about encrypted messages
	NoEncryptionWarning bool
	// RoomID where threads will be created
	RoomID string
	// Port for HTTP listener
	Port string
	// Prefix for honoroit commands
	Prefix string
	// LogLevel for logger
	LogLevel string
	// CacheSize max amount of items in cache
	CacheSize int

	// DB config
	DB DB

	// Auth Config
	Auth Auth

	// Redmine config
	Redmine Redmine

	// Monitoring config
	Monitoring Monitoring
}

// DB config
type DB struct {
	// DSN is a database connection string
	DSN string
	// Dialect of the db, allowed values: postgres, sqlite3
	Dialect string
}

type Auth struct {
	Metrics *echobasicauth.Auth
}

type Redmine struct {
	Host             string
	APIKey           string
	ProjectID        string
	TrackerID        int
	NewStatus        int
	InProgressStatus int
	DoneStatus       int
}

// Monitoring config
type Monitoring struct {
	SentryDSN            string
	SentrySampleRate     int
	HealthchecksURL      string
	HealthchecksUUID     string
	HealthchecksDuration time.Duration
}
