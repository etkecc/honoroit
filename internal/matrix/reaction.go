package matrix

import (
	"context"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) forwardReaction(ctx context.Context, evt *event.Event) {
	b.lock(evt.RoomID.String())
	defer b.unlock(evt.RoomID.String())

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
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot get event in the room")
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		linkpearl.ParseContent(sourceEvt, b.log)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(ctx, sourceEvt)
		if derr == nil {
			sourceEvt = decrypted
		}
	}

	linkpearl.ParseContent(sourceEvt, b.log)
	parentID := linkpearl.GetParent(sourceEvt)
	roomID, err := b.findRoomID(ctx, parentID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot find a suitable room to send reaction event")
		return
	}

	targetID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok { // message by operator doesn't contain metadata, try to find the same event in the customer's room
		targetEvent := b.lp.FindEventBy(ctx, roomID, map[string]string{"event_id": sourceID.String()})
		if targetEvent == nil {
			b.log.Error().Err(err).Str("sourceID", sourceID.String()).Msg("event found neither in operators nor in customers room")
			return
		}
		targetID = targetEvent.ID.String()
	}

	_, err = b.lp.GetClient().SendReaction(ctx, roomID, id.EventID(targetID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Msg("cannot send reaction")
	}
}

func (b *Bot) forwardReactionToThread(ctx context.Context, evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(ctx, evt.RoomID, sourceID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot get event in the room")
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		linkpearl.ParseContent(sourceEvt, b.log)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(ctx, sourceEvt)
		if derr == nil {
			sourceEvt = decrypted
		}
	}

	linkpearl.ParseContent(evt, b.log)
	operatorsRoomEventID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok {
		b.log.Error().Any("any", sourceEvt.Content.Raw["event_id"]).Msg("cannot cast source events event_id field")
		return
	}

	_, err = b.lp.GetClient().SendReaction(ctx, b.roomID, id.EventID(operatorsRoomEventID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send reaction")
	}
}
