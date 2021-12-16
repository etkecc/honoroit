package matrix

import (
	"strings"

	"maunium.net/go/mautrix/event"
)

func (b *Bot) parseCommand(message string) string {
	// ignore messages not prefixed with bot userID
	if !strings.HasPrefix(message, b.userID.String()) {
		return ""
	}

	message = strings.Replace(message, b.userID.String(), "", 1)
	message = strings.Replace(message, ":", "", 1)
	b.log.Debug("received a command: %s", message)
	return strings.TrimSpace(message)
}

func (b *Bot) runCommand(command string, evt *event.Event) {
	switch command {
	case "done", "complete", "close":
		b.closeRequest(evt)
	}
}

func (b *Bot) closeRequest(evt *event.Event) {
	b.log.Debug("closing a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.error(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I close your request.")
		return
	}

	roomID, err := b.findRoomID(relation.EventID)
	if err != nil {
		b.error(evt.RoomID, err.Error())
		return
	}

	_, err = b.api.SendMessageEvent(roomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Done,
	})
	if err != nil {
		b.error(evt.RoomID, err.Error())
		return
	}

	b.log.Debug("leaving room %s", roomID)
	_, err = b.api.LeaveRoom(roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(err.Error(), "M_FORBIDDEN") {
			b.error(evt.RoomID, "cannot leave the room %s after marking request as done: %v", roomID, err)
			return
		}
	}
}
