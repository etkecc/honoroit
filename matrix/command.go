package matrix

import (
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (b *Bot) parseCommand(message string) []string {
	command := []string{}
	// ignore not prefixied commands
	if !strings.HasPrefix(message, b.prefix) {
		return command
	}

	message = strings.TrimSpace(message[len(b.prefix)+1:])
	b.log.Debug("received a command: %s", message)

	command = strings.Split(message, " ")
	return command
}

func (b *Bot) readCommand(message string) string {
	command := b.parseCommand(message)
	if len(command) > 0 {
		return command[0]
	}
	return ""
}

func (b *Bot) runCommand(command string, evt *event.Event) {
	switch command {
	case "done", "complete", "close":
		b.closeRequest(evt)
	case "rename":
		b.renameRequest(evt)
	case "invite":
		b.inviteRequest(evt)
	default:
		b.help()
	}
}

func (b *Bot) renameRequest(evt *event.Event) {
	b.log.Debug("renaming a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I rename your request.")
		return
	}

	command := ""
	commandSlice := b.parseCommand(content.Body)
	if len(commandSlice) > 1 {
		command = strings.Join(commandSlice[1:], " ")
	}
	commandFormatted := ""
	commandSliceFormatted := b.parseCommand(content.FormattedBody)
	if len(commandSliceFormatted) > 1 {
		commandFormatted = strings.Join(commandSliceFormatted[1:], " ")
	}

	err := b.replace(relation.EventID, "", "", command, commandFormatted)
	if err != nil {
		b.Error(b.roomID, "cannot replace thread %s topic: %v", relation.EventID, err)
	}
}

func (b *Bot) closeRequest(evt *event.Event) {
	b.log.Debug("closing a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I close your request.")
		return
	}

	roomID, err := b.findRoomID(relation.EventID)
	if err != nil {
		b.Error(evt.RoomID, err.Error())
		return
	}

	_, err = b.lp.Send(roomID, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Done,
	})
	if err != nil {
		b.Error(evt.RoomID, err.Error())
	}
	timestamp := time.Now().UTC().Format("2006/01/02 15:04:05 MST")
	err = b.replace(relation.EventID, "[DONE] ", " ("+timestamp+")", "", "")
	if err != nil {
		b.Error(b.roomID, "cannot replace thread %s topic: %v", relation.EventID, err)
	}

	b.log.Debug("leaving room %s", roomID)
	_, err = b.lp.GetClient().LeaveRoom(roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(err.Error(), "M_FORBIDDEN") {
			b.Error(evt.RoomID, "cannot leave the room %s after marking request as done: %v", roomID, err)
		}
	}
	b.removeMapping("event_id", relation.EventID.String())
}

func (b *Bot) inviteRequest(evt *event.Event) {
	b.log.Debug("inviting the operator (%s) into customer room...", evt.Sender)
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I invite you.")
		return
	}

	roomID, err := b.findRoomID(relation.EventID)
	if err != nil {
		b.Error(evt.RoomID, err.Error())
		return
	}
	_, err = b.lp.GetClient().InviteUser(roomID, &mautrix.ReqInviteUser{
		Reason: "you asked it",
		UserID: evt.Sender,
	})

	if err != nil {
		b.Error(evt.RoomID, "cannot invite the operator (%s) into customer room %s: %v", evt.Sender, roomID, err)
	}
}

func (b *Bot) help() {
	b.log.Debug("help request")
	text := `Honoroit can perform following actions (note that all of them should be sent in a thread:

` + b.prefix + ` done - close the current request. Customer will receive a message about that and bot will leave the customer's room, thead topic will be prefixed with "[DONE]" suffixed with timestamp

` + b.prefix + ` rename TEXT - replaces thread topic text to the TEXT

` + b.prefix + ` invite - invite yourself into the customer's room
`
	_, err := b.lp.Send(b.roomID, &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    text,
	})
	if err != nil {
		b.Error(b.roomID, "cannot send help message: %v", err)
	}
}
