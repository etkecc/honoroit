package matrix

import (
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

	err = sourceEvt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot parse source event content")
		return
	}

	relatesTo := b.getRelatesTo(sourceEvt)
	if relatesTo == nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot parse source event relates_to")
		return
	}

	if relatesTo.EventID == "" {
		return
	}

	roomID, err := b.findRoomID(relatesTo.EventID)
	if err != nil {
		b.log.Error().Err(err).Str("sourceID", sourceID.String()).Str("roomID", evt.RoomID.String()).Msg("cannot find a suitable room to send reaction event")
		return
	}

	targetID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok { // message by operator doesn't contain metadata, try to find the same event in the customer's room
		targetEvent := b.findEventByAttr(roomID, "event_id", sourceID.String(), "")
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

	err = sourceEvt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot parse source event content")
		return
	}

	operatorsRoomEventID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok {
		b.log.Error().Msg("cannot cast source events event_id field")
		return
	}

	_, err = b.lp.GetClient().SendReaction(b.roomID, id.EventID(operatorsRoomEventID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send reaction")
	}
}
