package store

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"
)

// NOTE: functions in that file are for mautrix.Storer implementation
// ref: https://pkg.go.dev/maunium.net/go/mautrix#Storer

// SaveFilterID to DB
func (s *Store) SaveFilterID(userID id.UserID, filterID string) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}

	var insert string
	switch s.dialect {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO user_filter_ids VALUES (?, ?)"
	case "postgres":
		insert = "INSERT INTO user_filter_ids VALUES ($1, $2) ON CONFLICT DO NOTHING"
	}
	update := "UPDATE user_filter_ids SET filter_id = $1 WHERE user_id = $2"

	_, updateErr := tx.Exec(update, filterID, userID)
	if updateErr != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	_, insertErr := tx.Exec(insert, userID, filterID)
	if insertErr != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		// nolint // no need to check error here
		tx.Rollback()
	}
}

// LoadFilterID from DB
func (s *Store) LoadFilterID(userID id.UserID) string {
	query := "SELECT filter_id FROM user_filter_ids WHERE user_id = $1"
	row := s.db.QueryRow(query, userID)
	var filterID string
	if err := row.Scan(&filterID); err != nil {
		return ""
	}
	return filterID
}

// SaveNextBatch to DB
func (s *Store) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}

	var insert string
	switch s.dialect {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO user_batch_tokens VALUES (?, ?)"
	case "postgres":
		insert = "INSERT INTO user_batch_tokens VALUES ($1, $2) ON CONFLICT DO NOTHING"
	}
	update := "UPDATE user_batch_tokens SET next_batch_token = $1 WHERE user_id = $2"

	if _, err := tx.Exec(update, nextBatchToken, userID); err != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	if _, err := tx.Exec(insert, userID, nextBatchToken); err != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	// nolint // interface doesn't allow to return error
	tx.Commit()
}

// LoadNextBatch from DB
func (s *Store) LoadNextBatch(userID id.UserID) string {
	query := "SELECT next_batch_token FROM user_batch_tokens WHERE user_id = $1"
	row := s.db.QueryRow(query, userID)
	var batchToken string
	if err := row.Scan(&batchToken); err != nil {
		return ""
	}
	return batchToken
}

// SaveRoom to DB, not implemented
func (s *Store) SaveRoom(_ *mautrix.Room) {}

// LoadRoom from DB, not implemented
func (s *Store) LoadRoom(roomID id.RoomID) *mautrix.Room {
	return mautrix.NewRoom(roomID)
}
