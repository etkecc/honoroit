package matrix

import (
	"database/sql"
	"errors"

	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const mappingPrefix = "cc.etke.honoroit.mapping."

var (
	// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
	errNotMapped = errors.New("cannot find appropriate mapping")
	// errNotRelated returned if a message doesn't have relation (even recursive) to a thread
	errNotRelated = errors.New("cannot find appropriate thread")
)

func (b *Bot) findEventByAttr(roomID id.RoomID, attrName, attrValue, from string) *event.Event {
	resp, err := b.lp.GetClient().Messages(roomID, from, "", 'b', nil, 100)
	if err != nil {
		return nil
	}

	for _, msg := range resp.Chunk {
		// if it is encrypted
		if msg.Type == event.EventEncrypted {
			linkpearl.ParseContent(msg, b.log)
			decrypted, err := b.lp.GetClient().Crypto.Decrypt(msg)
			if err == nil {
				msg = decrypted
			}
		}

		if linkpearl.EventContains(msg, attrName, attrValue) {
			return msg
		}
	}

	if resp.End == "" { // nothing more
		return nil
	}

	return b.findEventByAttr(roomID, attrName, attrValue, resp.End)
}

func (b *Bot) findThread(evt *event.Event) (id.EventID, error) {
	threadID := linkpearl.GetParent(evt)

	v, ok := b.eventsCache.Get(string(threadID))
	if ok {
		eventID := v.(id.EventID)
		return eventID, nil
	}

	if threadID == evt.ID {
		return "", errNotRelated
	}

	b.eventsCache.Add(evt.ID, threadID)
	return threadID, nil
}

func (b *Bot) getMapping(id string) (string, error) {
	data, err := b.lp.GetAccountData(mappingPrefix + id)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", errNotMapped
	}

	v, ok := data["id"]
	if !ok {
		return "", errNotMapped
	}
	return v, nil
}

func (b *Bot) setMapping(from, to string) {
	err := b.lp.SetAccountData(mappingPrefix+from, map[string]string{"id": to})
	if err != nil {
		b.log.Error().Err(err).Msg("cannot set mapping")
	}
}

func (b *Bot) removeMapping(id string) {
	b.lp.SetAccountData(mappingPrefix+id, map[string]string{}) //nolint:errcheck // doesn't matter
}

// TODO remove after some time
// nolint:gocognit
func (b *Bot) migrateMappings() {
	query := "SELECT * FROM mappings"
	rows, err := b.lp.GetDB().Query(query)
	if err != nil {
		b.log.Info().Err(err).Msg("query of the old db mappings table failed, nothing to migrate")
		return
	}
	for rows.Next() {
		var roomID id.RoomID
		var email string
		var eventID id.EventID
		err := rows.Scan(&roomID, &email, &eventID)
		if err != nil && err != sql.ErrNoRows {
			b.log.Warn().Err(err).Msg("cannot load mapping from the old db")
			continue
		}

		if roomID == "" || eventID == "" {
			continue
		}

		b.setMapping(roomID.String(), eventID.String())
		b.setMapping(eventID.String(), roomID.String())

		// remove old mapping
		tx, err := b.lp.GetDB().Begin()
		if err != nil {
			continue
		}
		query := "DELETE FROM mappings WHERE room_id = $1"
		_, err = tx.Exec(query, roomID)
		if err != nil {
			tx.Rollback() //nolint:errcheck
			continue
		}

		tx.Commit() //nolint:errcheck
	}
}

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	roomID, err := b.getMapping(eventID.String())

	return id.RoomID(roomID), err
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	eventID, err := b.getMapping(roomID.String())
	if eventID == "" {
		return "", err
	}
	_, err = b.lp.GetClient().GetEvent(b.roomID, id.EventID(eventID))
	if err != nil {
		b.removeMapping(roomID.String())
		b.removeMapping(eventID)
		return "", errNotMapped
	}

	return id.EventID(eventID), nil
}
