// Package config was added to store cross-package structs and interfaces.
package config

import (
	"database/sql"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
)

// Config represents matrix config
type Config struct {
	// Homeserver url
	Homeserver string
	// Login is a localpart (honoroit - OK, @honoroit:example.com - wrong)
	Login string
	// Password for login/password auth only
	Password string

	// AutoLeave if true, linkpearl will automatically leave empty rooms
	AutoLeave bool

	// MaxRetries for operations like auto join
	MaxRetries int

	// NoEncryption disabled encryption support
	NoEncryption bool

	// LPLogger used for linkpearl's glue code
	LPLogger Logger
	// APILogger used for matrix CS API calls
	APILogger Logger
	// StoreLogger used for persistent store
	StoreLogger Logger
	// CryptoLogger used for OLM machine
	CryptoLogger Logger

	// DB object
	DB *sql.DB
	// Dialect of the DB: postgres, sqlite3
	Dialect string
}

// Logger implementation of crypto.Logger and mautrix.Logger
type Logger interface {
	crypto.Logger
	mautrix.WarnLogger

	Info(message string, args ...interface{})
}
