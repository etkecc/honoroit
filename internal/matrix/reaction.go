package matrix

import (
	"context"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) forwardReaction(ctx context.Context, evt *event.Event) {
	b.mu.Lock(evt.RoomID.String())
	defer b.mu.Unlock(evt.RoomID.String())

	if evt.RoomID == b.roomID {
		b.forwardReactionToCustomer(ctx, evt)
		return
	}
	b.forwardReactionToThread(ctx, evt)
}

func (b *Bot) forwardReactionToCustomer(ctx context.Context, evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(ctx, evt.RoomID, sourceID)
	if err != nil {
		b.log.Error("cannot get event in the room", "error", err, "sourceID", sourceID.String(), "roomID", evt.RoomID.String())
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		b.parseContent(sourceEvt)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(ctx, sourceEvt)
		if derr == nil {
			sourceEvt = decrypted
		}
	}

	b.parseContent(sourceEvt)
	parentID := linkpearl.GetParent(sourceEvt)
	roomID, err := b.findRoomID(ctx, parentID)
	if err != nil {
		b.log.Error("cannot find a suitable room to send reaction event", "error", err, "sourceID", sourceID.String(), "roomID", evt.RoomID.String())
		return
	}

	targetID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok { // message by operator doesn't contain metadata, try to find the same event in the customer's room
		targetEvent := b.lp.FindEventBy(ctx, roomID, map[string]string{"event_id": sourceID.String()})
		if targetEvent == nil {
			b.log.Error("event found neither in operators nor in customers room", "error", err, "sourceID", sourceID.String())
			return
		}
		targetID = targetEvent.ID.String()
	}

	_, err = b.lp.GetClient().SendReaction(ctx, roomID, id.EventID(targetID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error("cannot send reaction", "error", err, "sourceID", sourceID.String())
	}
}

func (b *Bot) forwardReactionToThread(ctx context.Context, evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(ctx, evt.RoomID, sourceID)
	if err != nil {
		b.log.Error("cannot get event in the room", "error", err, "sourceID", sourceID.String(), "roomID", evt.RoomID.String())
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		b.parseContent(sourceEvt)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(ctx, sourceEvt)
		if derr == nil {
			sourceEvt = decrypted
		}
	}

	b.parseContent(evt)
	operatorsRoomEventID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok {
		b.log.Error("cannot cast source events event_id field", "any", sourceEvt.Content.Raw["event_id"])
		return
	}

	_, err = b.lp.GetClient().SendReaction(ctx, b.roomID, id.EventID(operatorsRoomEventID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error("cannot send reaction", "error", err)
	}
}
