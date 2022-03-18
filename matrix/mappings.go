package matrix

import (
	"errors"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
var errNotMapped = errors.New("cannot find appropriate mapping")

// errNotRelated returned if a message doesn't have relation (even recursive) to a thread
var errNotRelated = errors.New("cannot find appropriate thread")

func (b *Bot) findThread(evt *event.Event) (id.EventID, error) {
	err := evt.Content.ParseRaw(event.EventMessage)
	if err != nil && err != event.ErrContentAlreadyParsed {
		return "", err
	}

	relation := evt.Content.AsMessage().RelatesTo
	// if relation is empty, consider that message as a thread root
	if relation == nil {
		return evt.ID, nil
	}

	threadID := b.getCache(relation.EventID)
	if threadID != "" {
		return threadID, nil
	}

	// If message relates to a thread - return thread root event ID
	if relation.Type == ThreadRelation || relation.Type == ThreadRelationOld {
		b.setCache(evt.ID, relation.EventID)
		return relation.EventID, nil
	}

	// If message is a reply-to, try to find a thread root
	if relation.Type == event.RelReply {
		return b.walkReplies(evt)
	}

	return "", errNotRelated
}

func (b *Bot) walkReplies(evt *event.Event) (id.EventID, error) {
	var err error
	relation := evt.Content.AsMessage().RelatesTo
	threadID := b.getCache(relation.EventID)
	if threadID != "" {
		return threadID, nil
	}

	evt, err = b.lp.GetClient().GetEvent(evt.RoomID, relation.EventID)
	if err != nil {
		return "", err
	}

	threadID, err = b.findThread(evt)
	if err != nil {
		return "", errNotRelated
	}

	b.setCache(evt.ID, threadID)
	return threadID, nil
}

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
