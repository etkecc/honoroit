package matrix

import (
	"errors"
	"strings"

	"maunium.net/go/mautrix/id"
)

const cacheMappings = "mappings"

type accountDataMappings struct {
	Rooms  map[id.RoomID]id.EventID `json:"rooms"`
	Events map[id.EventID]id.RoomID `json:"events"`
}

// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
var errNotMapped = errors.New("cannot find appropriate mapping")

func (b *Bot) getMappings() (*accountDataMappings, error) {
	var mappings *accountDataMappings
	b.log.Debug("trying to get mappings")

	cached := b.cache.Get(cacheMappings)
	if cached != nil {
		var ok bool
		mappings, ok = cached.(*accountDataMappings)
		if ok {
			b.log.Debug("got mappings from cache")
			return mappings, nil
		}
	}

	b.log.Debug("mappings not cached yet, trying to get them from account data")
	err := b.api.GetAccountData(accountDataRooms, &mappings)
	if err != nil {
		if strings.Contains(err.Error(), "M_NOT_FOUND") {
			return nil, nil
		}
		return nil, err
	}
	b.cache.Set(cacheMappings, mappings)
	return mappings, err
}

func (b *Bot) addMapping(roomID id.RoomID, eventID id.EventID) error {
	b.log.Debug("adding new mapping: %s<->%s", roomID, eventID)
	data, err := b.getMappings()
	if err != nil {
		return err
	}

	data.Rooms[roomID] = eventID
	data.Events[eventID] = roomID

	b.cache.Set(cacheMappings, data)
	err = b.api.SetAccountData(accountDataRooms, data)
	if err != nil {
		return err
	}

	return nil
}

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	b.log.Debug("trying to find room ID by eventID = %s", eventID)
	mappings, err := b.getMappings()
	if err != nil {
		return "", err
	}

	roomID, ok := mappings.Events[eventID]
	if !ok || roomID == "" {
		b.log.Debug("room not found in existing mappings")
		return "", errNotMapped
	}

	return roomID, nil
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	b.log.Debug("trying to find event ID by roomID = %s", roomID)
	mappings, err := b.getMappings()
	if err != nil {
		return "", err
	}

	eventID, ok := mappings.Rooms[roomID]
	if !ok || eventID == "" {
		b.log.Debug("event not found in existing mappings")
		return "", errNotMapped
	}

	return eventID, nil
}
