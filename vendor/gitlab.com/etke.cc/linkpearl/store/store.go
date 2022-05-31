// Package store implements crypto.Store, crypto.StateStore, mautrix.Storer and some additional "glue methods"
package store

import (
	"database/sql"

	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/linkpearl/config"
)

// Store for the matrix
type Store struct {
	db         *sql.DB
	dialect    string
	log        config.Logger
	encryption bool
	s          *crypto.SQLCryptoStore
}

// New store
func New(db *sql.DB, dialect string, log config.Logger) *Store {
	return &Store{
		db:      db,
		log:     log,
		dialect: dialect,
	}
}

// WithCrypto adds crypto store support
func (s *Store) WithCrypto(userID id.UserID, deviceID id.DeviceID, logger config.Logger) error {
	s.log.Debug("crypto store enabled")
	s.encryption = true
	s.s = crypto.NewSQLCryptoStore(
		s.db,
		s.dialect,
		userID.String(),
		deviceID,
		[]byte(userID),
		logger,
	)

	return s.s.CreateTables()
}

// GetDialect returns database dialect
func (s *Store) GetDialect() string {
	return s.dialect
}
