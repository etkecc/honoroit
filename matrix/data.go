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
			b.error(b.roomID, "sync rooms map error: "+err.Error())
		}
	}
}

func (b *Bot) loadRoomsMap() error {
	b.admu.Lock()
	defer b.admu.Unlock()
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
		err = b.loadRoomsMap()
	}

	return b.roomsMap, err
}

func (b *Bot) addRoomsMap(roomID id.RoomID, eventID id.EventID) error {
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

	// force reload mapping to avoid cache
	return nil
}

// findRoomID by eventID
func (b *Bot) findRoomID(eventID id.EventID) (id.RoomID, error) {
	rooms, err := b.getRoomsMap()
	if err != nil {
		return "", err
	}

	roomID, ok := rooms.Events[eventID]
	if !ok || roomID == "" {
		return "", errNotMapped
	}

	return roomID, nil
}

// findEventID by roomID
func (b *Bot) findEventID(roomID id.RoomID) (id.EventID, error) {
	rooms, err := b.getRoomsMap()
	if err != nil {
		return "", err
	}

	eventID, ok := rooms.Rooms[roomID]
	if !ok || eventID == "" {
		return "", errNotMapped
	}

	return eventID, nil
}
