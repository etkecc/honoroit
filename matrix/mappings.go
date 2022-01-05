package matrix

import (
	"errors"

	"maunium.net/go/mautrix/id"
)

// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
var errNotMapped = errors.New("cannot find appropriate mapping")

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	b.log.Debug("trying to find room ID by eventID = %s", eventID)
	roomID, _, _ := b.loadMapping("event_id", eventID.String())
	if roomID == "" {
		b.log.Debug("room not found in existing mappings")
		return "", errNotMapped
	}

	return roomID, nil
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	b.log.Debug("trying to find event ID by roomID = %s", roomID)
	_, _, eventID := b.loadMapping("room_id", roomID.String())
	if eventID == "" {
		b.log.Debug("room not found in existing mappings")
		return "", errNotMapped
	}
	return eventID, nil
}
