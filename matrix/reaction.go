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
		b.log.Error("cannot get event %s in the room %s: %v", sourceID, evt.RoomID, err)
		return
	}

	err = sourceEvt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.log.Error("cannot parse source event content: %v", err)
		return
	}

	relatesTo := b.getRelatesTo(sourceEvt)
	if relatesTo == nil {
		b.log.Error("cannot parse source event relates_to: %v", err)
		return
	}

	if relatesTo.EventID == "" {
		b.log.Error("cannot parse source event relates_to doesn't contain event id")
		return
	}

	roomID, err := b.findRoomID(relatesTo.EventID)
	if err != nil {
		b.log.Error("cannot find a suitable room to send reaction event: %v", err)
		return
	}

	targetID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok { // message by operator doesn't contain metadata, try to find the same event in the customer's room
		targetEvent := b.findEventByAttr(roomID, "event_id", sourceID.String(), "")
		if targetEvent == nil {
			b.log.Error("cannot find event %s neither in operators, nor in customer's room", sourceID)
			return
		}
		targetID = targetEvent.ID.String()
	}

	_, err = b.lp.GetClient().SendReaction(roomID, id.EventID(targetID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error("cannot send reaction: %v", err)
	}
}

func (b *Bot) forwardReactionToThread(evt *event.Event) {
	content := evt.Content.AsReaction()
	sourceID := content.GetRelatesTo().EventID
	sourceEvt, err := b.lp.GetClient().GetEvent(evt.RoomID, sourceID)
	if err != nil {
		b.log.Error("cannot get event %s in the room %s: %v", sourceID, evt.RoomID, err)
		return
	}

	err = sourceEvt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.log.Error("cannot parse source event content: %v", err)
		return
	}

	operatorsRoomEventID, ok := sourceEvt.Content.Raw["event_id"].(string)
	if !ok {
		b.log.Error("cannot cast source event's event_id field, raw: %v", sourceEvt.Content.Raw["event_id"])
		return
	}

	_, err = b.lp.GetClient().SendReaction(b.roomID, id.EventID(operatorsRoomEventID), content.RelatesTo.Key)
	if err != nil {
		b.log.Error("cannot send reaction: %v", err)
	}
}
