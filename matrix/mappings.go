package matrix

import (
	"errors"

	"maunium.net/go/mautrix/id"
)

type accountDataMappings struct {
	Rooms  map[id.RoomID]id.EventID `json:"rooms"`
	Events map[id.EventID]id.RoomID `json:"events"`
}

// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
var errNotMapped = errors.New("cannot find appropriate mapping")

// migrateMappings MIGRATION. TODO: remove
func (b *Bot) migrateMappings() {
	var data *accountDataMappings
	err := b.api.GetAccountData(accountDataRooms, &data)
	if err != nil {
		b.log.Error("cannot find account data mappings: %v", err)
		return
	}
	if data == nil {
		b.log.Debug("account data mappings not found")
	}

	for roomID, eventID := range data.Rooms {
		b.store.SaveMapping(roomID, "", eventID)
	}

	err = b.api.SetAccountData(accountDataRooms, struct{}{})
	if err != nil {
		b.log.Error("cannot wipe account data mappings: %v", err)
	}
}

func (b *Bot) addMapping(roomID id.RoomID, eventID id.EventID) {
	b.log.Debug("adding new mapping: %s<->%s", roomID, eventID)
	b.store.SaveMapping(roomID, "", eventID)
}

func (b *Bot) removeMapping(roomID id.RoomID, eventID id.EventID) {
	b.log.Debug("removing mapping %s<->%s...", roomID, eventID)
	b.store.RemoveMapping(roomID, "", eventID)
}

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	b.log.Debug("trying to find room ID by eventID = %s", eventID)
	roomID, _, _ := b.store.LoadMapping("", "", eventID)
	if roomID == "" {
		b.log.Debug("room not found in existing mappings")
		return "", errNotMapped
	}

	return roomID, nil
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	b.log.Debug("trying to find event ID by roomID = %s", roomID)
	_, _, eventID := b.store.LoadMapping(roomID, "", "")
	if eventID == "" {
		b.log.Debug("room not found in existing mappings")
		return "", errNotMapped
	}
	return eventID, nil
}
