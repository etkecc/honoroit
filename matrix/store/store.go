package store

import (
	"database/sql"

	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/logger"
)

// Store for the matrix
type Store struct {
	db      *sql.DB
	dialect string
	log     *logger.Logger
	s       *crypto.SQLCryptoStore
}

// New store
func New(db *sql.DB, dialect string, log *logger.Logger) *Store {
	return &Store{
		db:      db,
		log:     log,
		dialect: dialect,
	}
}

// WithCrypto adds crypto store support
func (s *Store) WithCrypto(userID id.UserID, deviceID id.DeviceID, logger crypto.Logger) error {
	s.log.Debug("crypto store enabled")
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
