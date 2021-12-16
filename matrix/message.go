package matrix

import (
	"fmt"
	"strings"
	"unsafe"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) error(roomID id.RoomID, message string) {
	fmt.Println("ERROR: ", message)
	// nolint // if something goes wrong here nobody can help...
	b.api.SendMessageEvent(roomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    "ERROR: " + message,
	})
}

func (b *Bot) handle(evt *event.Event) {
	// message sent from control room
	content := evt.Content.AsMessage()
	if unsafe.Sizeof(content) == 0 {
		b.error(evt.RoomID, "can't parse the message")
		return
	}

	if evt.RoomID == b.roomID {
		b.forwardToCustomer(evt, content)
		return
	}

	b.forwardToThread(evt, content)
}

func (b *Bot) startThread(roomID id.RoomID, userID id.UserID) (id.EventID, error) {
	eventID, err := b.findEventID(roomID)
	if err != nil && err != errNotMapped {
		b.error(b.roomID, fmt.Sprintf("user %s tried to send a message from room %s, but account data operation returned a error: %v", userID, roomID, err))
		b.error(roomID, "something is wrong. I already notified developers, they're fixing the issue. Please, try again later or use other contact methods.")
		return "", err
	}

	if eventID != "" {
		return eventID, nil
	}

	content := &event.MessageEventContent{
		MsgType: "m.message",
		Body:    "Request by " + userID.String() + " in " + roomID.String(),
	}

	resp, err := b.api.SendMessageEvent(b.roomID, event.EventMessage, content)
	if err != nil {
		b.error(b.roomID, fmt.Sprintf("user %s tried to send a message from room %s, but creation of a thread failed: %v", userID, roomID, err))
		return "", err
	}

	err = b.addRoomsMap(roomID, resp.EventID)
	if err != nil && err != errNotMapped {
		b.error(b.roomID, fmt.Sprintf("user %s tried to send a message from room %s, but account data operation failed: %v", userID, roomID, err))
	}

	return resp.EventID, nil
}

func (b *Bot) forwardToCustomer(evt *event.Event, content *event.MessageEventContent) {
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

	content.RelatesTo = nil
	_, err = b.api.SendMessageEvent(roomID, evt.Type, content)
	if err != nil {
		b.error(evt.RoomID, err.Error())
	}
}

func (b *Bot) forwardToThread(evt *event.Event, content *event.MessageEventContent) {
	eventID, err := b.startThread(evt.RoomID, evt.Sender)
	if err != nil {
		b.error(evt.RoomID, "something is wrong. I already notified developers, they're fixing the issue. Please, try again later or use other contact methods.")
		return
	}

	content.RelatesTo = &event.RelatesTo{
		Type:    ThreadRelation,
		EventID: eventID,
	}

	// sanitization to avoid calls of `/` commands
	if strings.HasPrefix(content.Body, "/") {
		content.Body = strings.Replace(content.Body, "/", "_/", 1)
	}
	if strings.HasPrefix(content.FormattedBody, "/") {
		content.FormattedBody = strings.Replace(content.FormattedBody, "/", "_/", 1)
	}

	_, err = b.api.SendMessageEvent(b.roomID, event.EventMessage, content)
	if err != nil {
		b.error(b.roomID, fmt.Sprintf("user %s tried to send a message from room %s, but creation of a thread failed: %v", evt.Sender, evt.RoomID, err))
		b.error(evt.RoomID, "something is wrong. I already notified developers, they're fixing the issue. Please, try again later or use other contact methods.")
	}
}
