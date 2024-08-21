package matrix

import (
	"context"
	"errors"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const (
	mappingPrefix        = "cc.etke.honoroit.mapping."
	mappingPrefixRedmine = "cc.etke.honoroit.redmine."
)

var (
	// errNotMapped returned if roomID or eventID doesn't exist in room<->event map (yet)
	errNotMapped = errors.New("cannot find appropriate mapping")
	// errNotRelated returned if a message doesn't have relation (even recursive) to a thread
	errNotRelated = errors.New("cannot find appropriate thread")
)

func (b *Bot) findThread(evt *event.Event) (id.EventID, error) {
	threadID := linkpearl.GetParent(evt)

	v, ok := b.eventsCache.Get(threadID)
	if ok {
		return v, nil
	}

	if threadID == evt.ID {
		return "", errNotRelated
	}

	b.eventsCache.Add(evt.ID, threadID)
	return threadID, nil
}

func (b *Bot) getMapping(ctx context.Context, identifier string) (string, error) {
	data, err := b.lp.GetAccountData(ctx, mappingPrefix+identifier)
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

func (b *Bot) setMapping(ctx context.Context, from, to string) {
	err := b.lp.SetAccountData(ctx, mappingPrefix+from, map[string]string{"id": to})
	if err != nil {
		b.log.Error().Err(err).Msg("cannot set mapping")
	}
}

func (b *Bot) removeMapping(ctx context.Context, identifier string) {
	b.lp.SetAccountData(ctx, mappingPrefix+identifier, map[string]string{}) //nolint:errcheck // doesn't matter
}

func (b *Bot) getRedmineMapping(ctx context.Context, identifier string) (string, error) {
	data, err := b.lp.GetAccountData(ctx, mappingPrefixRedmine+identifier)
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

func (b *Bot) setRedmineMapping(ctx context.Context, from, to string) {
	err := b.lp.SetAccountData(ctx, mappingPrefixRedmine+from, map[string]string{"id": to})
	if err != nil {
		b.log.Error().Err(err).Msg("cannot set mapping")
	}
}

func (b *Bot) removeRedmineMapping(ctx context.Context, identifier string) {
	b.lp.SetAccountData(ctx, mappingPrefixRedmine+identifier, map[string]string{}) //nolint:errcheck // doesn't matter
}

// findRoomID by eventID
func (b *Bot) findRoomID(ctx context.Context, eventID id.EventID) (id.RoomID, error) {
	roomID, err := b.getMapping(ctx, eventID.String())

	return id.RoomID(roomID), err
}

// findEventID by roomID
func (b *Bot) findEventID(ctx context.Context, roomID id.RoomID) (id.EventID, error) {
	eventID, err := b.getMapping(ctx, roomID.String())
	if eventID == "" {
		return "", err
	}
	_, err = b.lp.GetClient().GetEvent(ctx, b.roomID, id.EventID(eventID))
	if err != nil {
		b.removeMapping(ctx, roomID.String())
		b.removeMapping(ctx, eventID)
		return "", errNotMapped
	}

	return id.EventID(eventID), nil
}
