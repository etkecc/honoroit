package store

import (
	"database/sql"

	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/id"
)

// Store for the matrix
type Store struct {
	s      *crypto.SQLCryptoStore
	userID id.UserID
	// nextBatchToken MIGRATION. TODO: remove
	nextBatchToken string
}

// Config the store
type Config struct {
	// DB - sql dirver object
	DB *sql.DB
	// Dialect - one of: postgres, sqlite3
	Dialect string
	// UserID to use with
	UserID id.UserID
	// DeviceID to use with
	DeviceID id.DeviceID
	// Logger object
	Logger crypto.Logger

	// NextBatchToken MIGRATION. TODO: remove
	NextBatchToken string
}

// New store
func New(cfg *Config) *Store {
	cs := crypto.NewSQLCryptoStore(
		cfg.DB,
		cfg.Dialect,
		cfg.UserID.String(),
		cfg.DeviceID,
		[]byte(cfg.UserID),
		cfg.Logger,
	)

	return &Store{
		s:              cs,
		userID:         cfg.UserID,
		nextBatchToken: cfg.NextBatchToken,
	}
}
