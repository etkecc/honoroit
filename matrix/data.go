package matrix

import (
	"errors"
	"strings"
	"time"

	"maunium.net/go/mautrix/id"
)

type accountDataRoomsMap struct {
	Rooms  map[id.RoomID]id.EventID `json:"rooms"`
	Events map[id.EventID]id.RoomID `json:"events"`
}

// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
var errNotMapped = errors.New("cannot find appropriate mapping")

func (b *Bot) syncRoomsMap() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		if err := b.loadRoomsMap(); err != nil {
			b.Error(b.roomID, "sync rooms map error: %v", err)
		}
	}
}

func (b *Bot) loadRoomsMap() error {
	b.admu.Lock()
	defer b.admu.Unlock()

	b.log.Debug("refreshing rooms<->events map from account data")
	b.roomsMap = &accountDataRoomsMap{
		Rooms:  make(map[id.RoomID]id.EventID),
		Events: make(map[id.EventID]id.RoomID),
	}

	err := b.api.GetAccountData(accountDataRooms, &b.roomsMap)
	if err != nil && strings.Contains(err.Error(), "M_NOT_FOUND") {
		return nil
	}

	return err
}

func (b *Bot) getRoomsMap() (*accountDataRoomsMap, error) {
	var err error
	if b.roomsMap == nil || len(b.roomsMap.Rooms) == 0 {
		b.log.Debug("no rooms<->events map in memory cache, requesting from account data")
		err = b.loadRoomsMap()
	}

	return b.roomsMap, err
}

func (b *Bot) addRoomsMap(roomID id.RoomID, eventID id.EventID) error {
	b.log.Debug("adding new rooms<->events map item: %s<->%s", roomID, eventID)
	data, err := b.getRoomsMap()
	if err != nil {
		return err
	}

	b.admu.Lock()
	defer b.admu.Unlock()

	data.Rooms[roomID] = eventID
	data.Events[eventID] = roomID

	err = b.api.SetAccountData(accountDataRooms, data)
	if err != nil {
		return err
	}

	return nil
}

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	b.log.Debug("trying to find room ID by eventID = %s", eventID)
	rooms, err := b.getRoomsMap()
	if err != nil {
		return "", err
	}

	roomID, ok := rooms.Events[eventID]
	if !ok || roomID == "" {
		b.log.Debug("room not found in existing rooms<->events map")
		return "", errNotMapped
	}

	return roomID, nil
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	b.log.Debug("trying to find event ID by roomID = %s", roomID)
	rooms, err := b.getRoomsMap()
	if err != nil {
		return "", err
	}

	eventID, ok := rooms.Rooms[roomID]
	if !ok || eventID == "" {
		b.log.Debug("event not found in existing rooms<->events map")
		return "", errNotMapped
	}

	return eventID, nil
}
