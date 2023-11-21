package matrix

import (
	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) forwardReaction(evt *event.Event) {
	b.lock(evt.RoomID.String())
	defer b.unlock(evt.RoomID.String())

	if evt.RoomID == b.roomID {
		b.forwardReactionToCustomer(evt)
		return
	}
	b.forwardReactionToThread(evt)
}

func (b *Bot) forwardReactionToCustomer(evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(evt.RoomID, sourceID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot get event in the room")
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		linkpearl.ParseContent(sourceEvt, b.log)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(sourceEvt)
		if derr == nil {
			sourceEvt = decrypted
		}
	}

	linkpearl.ParseContent(sourceEvt, b.log)
	parentID := linkpearl.GetParent(sourceEvt)
	roomID, err := b.findRoomID(parentID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot find a suitable room to send reaction event")
		return
	}

	targetID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok { // message by operator doesn't contain metadata, try to find the same event in the customer's room
		targetEvent := b.lp.FindEventBy(roomID, "event_id", sourceID.String())
		if targetEvent == nil {
			b.log.Error().Err(err).Str("sourceID", sourceID.String()).Msg("event found neither in operators nor in customers room")
			return
		}
		targetID = targetEvent.ID.String()
	}

	_, err = b.lp.GetClient().SendReaction(roomID, id.EventID(targetID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Msg("cannot send reaction")
	}
}

func (b *Bot) forwardReactionToThread(evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(evt.RoomID, sourceID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot get event in the room")
		return
	}

	// if it is encrypted
	if sourceEvt.Type == event.EventEncrypted {
		linkpearl.ParseContent(sourceEvt, b.log)
		decrypted, derr := b.lp.GetClient().Crypto.Decrypt(sourceEvt)
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

	_, err = b.lp.GetClient().SendReaction(b.roomID, id.EventID(operatorsRoomEventID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send reaction")
	}
}
