package store

import (
	"encoding/json"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// NOTE: functions in that file are for crypto.StateStore implementation
// ref: https://pkg.go.dev/maunium.net/go/mautrix/crypto#StateStore

// IsEncrypted returns whether a room is encrypted.
func (s *Store) IsEncrypted(roomID id.RoomID) bool {
	return s.GetEncryptionEvent(roomID) != nil
}

// GetEncryptionEvent returns the encryption event's content for an encrypted room.
func (s *Store) GetEncryptionEvent(roomID id.RoomID) *event.EncryptionEventContent {
	query := "SELECT encryption_event FROM rooms WHERE room_id = $1"
	row := s.db.QueryRow(query, roomID)

	var encryptionEventJSON []byte
	if err := row.Scan(&encryptionEventJSON); err != nil {
		return nil
	}
	var encryptionEvent event.EncryptionEventContent
	if err := json.Unmarshal(encryptionEventJSON, &encryptionEvent); err != nil {
		return nil
	}

	return &encryptionEvent
}

// SetEncryptionEvent creates or updates room's encryption event info
func (s *Store) SetEncryptionEvent(evt *event.Event) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}

	var insert string
	switch s.dialect {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO rooms VALUES (?, ?)"
	case "postgres":
		insert = "INSERT INTO rooms VALUES ($1, $2) ON CONFLICT DO NOTHING"
	}
	update := "UPDATE rooms SET encryption_event = $1 WHERE room_id = $2"

	var encryptionEventJSON []byte
	encryptionEventJSON, err = json.Marshal(evt)
	if err != nil {
		encryptionEventJSON = nil
	}

	_, err = tx.Exec(update, encryptionEventJSON, evt.RoomID)
	if err != nil {
		// nolint // we already have err to return
		tx.Rollback()
		return
	}

	_, err = tx.Exec(insert, evt.RoomID, encryptionEventJSON)
	if err != nil {
		// nolint // interface doesn't allow to return error
		tx.Rollback()
		return
	}

	// nolint // interface doesn't allow to return error
	tx.Commit()
}

// SetMembership saves room members
func (s *Store) SetMembership(evt *event.Event) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}

	var insert string
	switch s.dialect {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO room_members VALUES (?, ?)"
	case "postgres":
		insert = "INSERT INTO room_members VALUES ($1, $2) ON CONFLICT DO NOTHING"
	}
	del := "DELETE FROM room_members WHERE room_id = $1 AND user_id = $2"

	membershipEvent := evt.Content.AsMember()
	if membershipEvent.Membership.IsInviteOrJoin() {
		_, err := tx.Exec(insert, evt.RoomID, evt.GetStateKey())
		if err != nil {
			// nolint // interface doesn't allow to return error
			tx.Rollback()
			return
		}
	} else {
		_, err := tx.Exec(del, evt.RoomID, evt.GetStateKey())
		if err != nil {
			// nolint // interface doesn't allow to return error
			tx.Rollback()
			return
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		// nolint // interface doesn't allow to return error
		tx.Rollback()
	}
}

// GetRoomMembers ...
func (s *Store) GetRoomMembers(roomID id.RoomID) []id.UserID {
	query := "SELECT user_id FROM room_members WHERE room_id = $1"
	rows, err := s.db.Query(query, roomID)
	users := make([]id.UserID, 0)
	if err != nil {
		return users
	}
	defer rows.Close()

	var userID id.UserID
	for rows.Next() {
		if err := rows.Scan(&userID); err == nil {
			users = append(users, userID)
		}
	}
	return users
}

// FindSharedRooms returns the encrypted rooms that another user is also in for a user ID.
func (s *Store) FindSharedRooms(userID id.UserID) []id.RoomID {
	query := "SELECT room_id FROM room_members WHERE user_id = $1"
	rows, queryErr := s.db.Query(query, userID)
	rooms := make([]id.RoomID, 0)
	if queryErr != nil {
		return rooms
	}
	defer rows.Close()

	var roomID id.RoomID
	for rows.Next() {
		scanErr := rows.Scan(&roomID)
		if scanErr != nil {
			continue
		}
		rooms = append(rooms, roomID)
	}

	return rooms
}

// SaveSession to DB
func (s *Store) SaveSession(userID id.UserID, deviceID id.DeviceID, accessToken string) {
	tx, err := s.db.Begin()
	if err != nil {
		return
	}

	var insert string
	switch s.dialect {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO session VALUES (?, ?, ?)"
	case "postgres":
		insert = "INSERT INTO session VALUES ($1, $2, $3) ON CONFLICT DO NOTHING"
	}
	update := "UPDATE session SET access_token = $1, device_id = $2 WHERE user_id = $3"

	if _, err := tx.Exec(update, accessToken, deviceID, userID); err != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	if _, err := tx.Exec(insert, userID, deviceID, accessToken); err != nil {
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	// nolint // interface doesn't allow to return error
	tx.Commit()
}

// LoadSession from DB (user ID, device ID, access token)
func (s *Store) LoadSession() (id.UserID, id.DeviceID, string) {
	row := s.db.QueryRow("SELECT * FROM session LIMIT 1")
	var userID id.UserID
	var deviceID id.DeviceID
	var accessToken string
	if err := row.Scan(&userID, &deviceID, &accessToken); err != nil {
		return "", "", ""
	}
	return userID, deviceID, accessToken
}
