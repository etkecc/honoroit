package store

var migrations = []string{
	`
		CREATE TABLE IF NOT EXISTS user_filter_ids (
			user_id    VARCHAR(255) PRIMARY KEY,
			filter_id  VARCHAR(255)
		)
		`,
	`
		CREATE TABLE IF NOT EXISTS user_batch_tokens (
			user_id           VARCHAR(255) PRIMARY KEY,
			next_batch_token  VARCHAR(255)
		)
		`,
	`
		CREATE TABLE IF NOT EXISTS rooms (
			room_id           VARCHAR(255) PRIMARY KEY,
			encryption_event  VARCHAR(65535) NULL
		)
		`,
	`
		CREATE TABLE IF NOT EXISTS room_members (
			room_id  VARCHAR(255),
			user_id  VARCHAR(255),
			PRIMARY KEY (room_id, user_id)
		)
		`,
}

// CreateTables applies all the pending database migrations.
func (s *Store) CreateTables() error {
	if err := s.s.CreateTables(); err != nil {
		return err
	}

	return s.migrate()
}

func (s *Store) migrate() error {
	tx, beginErr := s.s.DB.Begin()
	if beginErr != nil {
		return beginErr
	}

	for _, query := range migrations {
		_, execErr := tx.Exec(query)
		if execErr != nil {
			// nolint // we already have the execErr to return
			tx.Rollback()
			return execErr
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		// nolint // we already have the commitErr to return
		tx.Rollback()
		return commitErr
	}

	// MIGRATION. TODO: remove
	if s.nextBatchToken != "" {
		s.SaveNextBatch(s.userID, s.nextBatchToken)
	}

	return nil
}
