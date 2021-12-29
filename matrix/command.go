package matrix

import (
	"strings"
	"time"

	"maunium.net/go/mautrix/event"
)

func (b *Bot) parseCommand(message string) []string {
	command := []string{}
	// in some cases, localpart or MXID may be sent, so let's handle both
	userID := b.api.UserID.String()
	// nolint // we don't need to verify user id, just get the localpart
	localpart, _, _ := b.api.UserID.Parse()
	// ignore messages not prefixed with bot mention
	if !strings.HasPrefix(message, userID) && !strings.HasPrefix(message, localpart) && !strings.HasPrefix(message, b.name) {
		return command
	}

	b.log.Debug("received a command: %s", message)
	message = strings.Replace(message, userID, "", 1)
	message = strings.Replace(message, b.name, "", 1)
	message = strings.Replace(message, localpart, "", 1)
	delimiter := strings.Index(message, ":")
	message = strings.TrimSpace(message[delimiter+1:])
	b.log.Debug("parsed a command: %s", message)

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

	_, err = b.send(roomID, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Done,
	})
	if err != nil {
		b.Error(evt.RoomID, err.Error())
		return
	}
	timestamp := time.Now().UTC().Format("2006/01/02 15:04:05 MST")
	err = b.replace(relation.EventID, "[DONE] ", " ("+timestamp+")", "", "")
	if err != nil {
		b.Error(b.roomID, "cannot replace thread %s topic: %v", relation.EventID, err)
	}

	b.log.Debug("leaving room %s", roomID)
	_, err = b.api.LeaveRoom(roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(err.Error(), "M_FORBIDDEN") {
			b.Error(evt.RoomID, "cannot leave the room %s after marking request as done: %v", roomID, err)
			return
		}
	}
}
