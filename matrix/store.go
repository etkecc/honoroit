package matrix

import (
	"database/sql"

	"maunium.net/go/mautrix/id"
)

var migrations = []string{
	`
		CREATE TABLE IF NOT EXISTS mappings (
			room_id VARCHAR(255),
			email VARCHAR(255),
			event_id VARCHAR(255)
		)
		`,
}

func (b *Bot) migrate() error {
	b.log.Debug("migrating database...")
	tx, beginErr := b.lp.GetDB().Begin()
	if beginErr != nil {
		b.log.Error("cannot begin transaction: %v", beginErr)
		return beginErr
	}

	for _, query := range migrations {
		_, execErr := tx.Exec(query)
		if execErr != nil {
			b.log.Error("cannot apply migration: %v", execErr)
			// nolint // we already have the execErr to return
			tx.Rollback()
			return execErr
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		b.log.Error("cannot commit transaction: %v", commitErr)
		// nolint // we already have the commitErr to return
		tx.Rollback()
		return commitErr
	}

	return nil
}

func (b *Bot) saveMapping(roomID id.RoomID, email string, eventID id.EventID) {
	b.log.Debug("saving mapping of %s/%s/%s", roomID, email, eventID)
	tx, err := b.lp.GetDB().Begin()
	if err != nil {
		b.log.Error("cannot begin transaction: %v", err)
		return
	}

	var insert string
	switch b.lp.GetStore().GetDialect() {
	case "sqlite3":
		insert = "INSERT OR IGNORE INTO mappings VALUES (?, ?, ?)"
	case "postgres":
		insert = "INSERT INTO mappings VALUES ($1, $2, $3) ON CONFLICT DO NOTHING"
	}

	if _, err = tx.Exec(insert, roomID, email, eventID); err != nil {
		b.log.Error("cannot insert mapping: %v", err)
		// nolint // no need to check error here
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		b.log.Error("cannot commit transaction: %v", err)
	}
}

// nolint // email will be used later
func (b *Bot) loadMapping(field, value string) (id.RoomID, string, id.EventID) {
	b.log.Debug("loading mapping of with %s = %s", field, value)
	query := "SELECT * FROM mappings WHERE " + field + " = $1"
	row := b.lp.GetDB().QueryRow(query, value)
	var roomID id.RoomID
	var email string
	var eventID id.EventID
	err := row.Scan(&roomID, &email, &eventID)
	if err != nil && err != sql.ErrNoRows {
		b.log.Error("cannot load mapping: %v", err)
		return "", "", ""
	}

	return roomID, email, eventID
}

func (b *Bot) removeMapping(field, value string) {
	b.log.Debug("removing mapping of %s = %s", field, value)
	tx, err := b.lp.GetDB().Begin()
	if err != nil {
		b.log.Error("cannot begin transaction: %v", err)
		return
	}
	query := "DELETE FROM mappings WHERE " + field + " = $1"
	_, err = tx.Exec(query, value)
	if err != nil {
		b.log.Error("cannot delete mapping: %v", err)
		// nolint // no need
		tx.Rollback()
		return
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		b.log.Error("cannot commit transaction: %v", commitErr)
		// nolint // no need
		tx.Rollback()
		return
	}
}
