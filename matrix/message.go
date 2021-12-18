package matrix

import (
	"unsafe"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) error(roomID id.RoomID, message string, args ...interface{}) {
	b.log.Error(message, args...)
	// nolint // if something goes wrong here nobody can help...
	b.api.SendMessageEvent(roomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    "ERROR: " + message,
	})
}

func (b *Bot) greetings(roomID id.RoomID, userID id.UserID) {
	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Greetings,
	}
	if _, err := b.api.SendMessageEvent(roomID, event.EventMessage, content); err != nil {
		b.error(b.roomID, "cannot send a greetings message to the user %s in the room %s: %v", userID, roomID, err)
	}
}

func (b *Bot) typing(roomID id.RoomID, typing bool) {
	_, err := b.api.UserTyping(roomID, typing, TypingTimeout)
	if err != nil {
		b.log.Error("cannot send typing = %t status to the room %s: %v", typing, roomID, err)
	}
}

func (b *Bot) handle(evt *event.Event) {
	err := b.api.MarkRead(evt.RoomID, evt.ID)
	if err != nil {
		b.error(b.roomID, "cannot send mark event to the room %s: %v", evt.RoomID, err)
	}

	content := evt.Content.AsMessage()
	if unsafe.Sizeof(content) == 0 {
		b.error(evt.RoomID, "cannot parse the message")
		return
	}

	// message sent by client
	if evt.RoomID != b.roomID {
		b.typing(evt.RoomID, true)
		defer b.typing(evt.RoomID, false)

		b.forwardToThread(evt, content)
		return
	}

	// message sent from threads room
	// special command
	if command := b.readCommand(content.Body); command != "" {
		b.runCommand(command, evt)
		return
	}

	// not a command, but a message
	b.forwardToCustomer(evt, content)
}

func (b *Bot) replace(eventID id.EventID, prefix string, suffix string, body string, formattedBody string) error {
	evt, err := b.api.GetEvent(b.roomID, eventID)
	if err != nil {
		b.error(b.roomID, "cannot find event %s: %v", eventID, err)
		return err
	}

	err = evt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.error(b.roomID, "cannot parse thread topic event %s: %v", eventID, err)
		return err
	}
	content := evt.Content.AsMessage()
	if body == "" {
		body = content.Body
	}
	if formattedBody == "" {
		formattedBody = content.FormattedBody
	}

	body = prefix + body + suffix
	formattedBody = prefix + formattedBody + suffix

	content.Body = " * " + body
	content.FormattedBody = " * " + formattedBody
	content.NewContent = &event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          body,
		FormattedBody: formattedBody,
	}
	content.RelatesTo = &event.RelatesTo{
		EventID: eventID,
		Type:    event.RelReplace,
	}

	b.log.Debug("replacing thread topic event")
	_, err = b.api.SendMessageEvent(b.roomID, event.EventMessage, content)
	return err
}

func (b *Bot) startThread(roomID id.RoomID, userID id.UserID) (id.EventID, error) {
	b.log.Debug("starting new thread for %s request from %s", userID, roomID)
	eventID, err := b.findEventID(roomID)
	if err != nil && err != errNotMapped {
		b.error(b.roomID, "user %s tried to send a message from room %s, but account data operation returned a error: %v", userID, roomID, err)
		b.error(roomID, b.txt.Error)
		return "", err
	}

	if eventID != "" {
		b.log.Debug("thread for %s request from %s exists, returning eventID %s instead", userID, roomID, eventID)
		return eventID, nil
	}

	content := &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    "Request by " + userID.String() + " in " + roomID.String(),
	}

	resp, err := b.api.SendMessageEvent(b.roomID, event.EventMessage, content)
	if err != nil {
		b.error(b.roomID, "user %s tried to send a message from room %s, but creation of a thread failed: %v", userID, roomID, err)
		return "", err
	}

	err = b.addRoomsMap(roomID, resp.EventID)
	if err != nil && err != errNotMapped {
		b.error(b.roomID, "user %s tried to send a message from room %s, but account data operation failed: %v", userID, roomID, err)
	}

	b.greetings(roomID, userID)
	return resp.EventID, nil
}

func (b *Bot) forwardToCustomer(evt *event.Event, content *event.MessageEventContent) {
	b.log.Debug("forwaring the message to a customer room")
	relation := content.RelatesTo
	if relation == nil {
		b.error(evt.RoomID, "the message doesn't relate to any thread, so I don't know where to forward it.")
		return
	}

	roomID, err := b.findRoomID(relation.EventID)
	if err != nil {
		b.error(evt.RoomID, err.Error())
		return
	}

	b.typing(roomID, true)
	defer b.typing(roomID, false)

	content.RelatesTo = nil
	_, err = b.api.SendMessageEvent(roomID, evt.Type, content)
	if err != nil {
		b.error(evt.RoomID, err.Error())
	}
}

func (b *Bot) forwardToThread(evt *event.Event, content *event.MessageEventContent) {
	b.log.Debug("forwaring a message from customer to the threads rooms")
	eventID, err := b.startThread(evt.RoomID, evt.Sender)
	if err != nil {
		b.error(evt.RoomID, b.txt.Error)
		return
	}

	content.RelatesTo = &event.RelatesTo{
		Type:    ThreadRelation,
		EventID: eventID,
	}

	_, err = b.api.SendMessageEvent(b.roomID, event.EventMessage, content)
	if err != nil {
		b.error(b.roomID, "user %s tried to send a message from room %s, but creation of a thread failed: %v", evt.Sender, evt.RoomID, err)
		b.error(evt.RoomID, b.txt.Error)
	}
}
