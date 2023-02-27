package matrix

import (
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) parseCommand(message string) []string {
	if message == "" {
		return nil
	}

	if !strings.HasPrefix(message, b.prefix) {
		return nil
	}

	message = strings.Replace(message, b.prefix, "", 1)
	b.log.Debug("received a command: %s", message)
	return strings.Split(strings.TrimSpace(message), " ")
}

func (b *Bot) readCommand(message string) string {
	command := b.parseCommand(strings.TrimSpace(message))
	if len(command) > 0 {
		return command[0]
	}
	return ""
}

func (b *Bot) runCommand(command string, evt *event.Event, hub *sentry.Hub) {
	switch command {
	case "done", "complete", "close":
		b.closeRequest(evt, hub)
	case "rename":
		b.renameRequest(evt, hub)
	case "invite":
		b.inviteRequest(evt, hub)
	case "start":
		b.startRequest(evt, hub)
	case "note":
		// do nothing
		return
	default:
		b.help(evt, hub)
	}
}

func (b *Bot) renameRequest(evt *event.Event, hub *sentry.Hub) {
	b.log.Debug("renaming a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "the message doesn't relate to any thread, so I don't know how can I rename your request.")
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "cannot find a thread: %v", err)
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

	err = b.replace(threadID, hub, "", "", command, commandFormatted)
	if err != nil {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot replace thread %s topic: %v", threadID, err)
	}
}

func (b *Bot) closeRequest(evt *event.Event, hub *sentry.Hub) {
	b.log.Debug("closing a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "the message doesn't relate to any thread, so I don't know how can I close your request.")
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "cannot find a thread: %v", err)
		return
	}

	threadEvt, err := b.lp.GetClient().GetEvent(b.roomID, threadID)
	if err != nil {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot find thread event %s: %v", threadID, err)
		return
	}

	roomID, err := b.findRoomID(threadID)
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, err.Error())
		return
	}

	_, err = b.lp.Send(roomID, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Done,
	})
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, err.Error())
	}

	var oldbody string
	if threadEvt.Content.AsMessage() != nil {
		oldbody = strings.Replace(threadEvt.Content.AsMessage().Body, b.txt.PrefixOpen, "", 1)
	}
	timestamp := time.Now().UTC().Format("2006/01/02 15:04:05 MST")
	err = b.replace(threadID, hub, b.txt.PrefixDone+" ", oldbody+" ("+timestamp+")", "", "")
	if err != nil {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot replace thread %s topic: %v", threadID, err)
	}

	b.log.Debug("leaving room %s", roomID)
	_, err = b.lp.GetClient().LeaveRoom(roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(err.Error(), "M_FORBIDDEN") {
			b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "cannot leave the room %s after marking request as done: %v", roomID, err)
		}
	}
	b.removeMapping("event_id", threadID.String())
}

func (b *Bot) inviteRequest(evt *event.Event, hub *sentry.Hub) {
	b.log.Debug("inviting the operator (%s) into customer room...", evt.Sender)
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "the message doesn't relate to any thread, so I don't know how can I invite you.")
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "cannot find a thread: %v", err)
		return
	}

	roomID, err := b.findRoomID(threadID)
	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, err.Error())
		return
	}
	_, err = b.lp.GetClient().InviteUser(roomID, &mautrix.ReqInviteUser{
		Reason: "you asked it",
		UserID: evt.Sender,
	})

	if err != nil {
		b.Error(evt.RoomID, b.getRelatesTo(evt), hub, "cannot invite the operator (%s) into customer room %s: %v", evt.Sender, roomID, err)
	}
}

func (b *Bot) startRequest(evt *event.Event, hub *sentry.Hub) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) < 2 {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot start a new matrix room - MXID is not specified")
		return
	}
	userID := id.UserID(command[1])
	req := &mautrix.ReqCreateRoom{
		Invite:   []id.UserID{userID},
		Preset:   "trusted_private_chat",
		IsDirect: true,
	}
	if b.lp.GetMachine() != nil {
		req.InitialState = []*event.Event{
			{
				Type: event.StateEncryption,
				Content: event.Content{
					Parsed: event.EncryptionEventContent{
						Algorithm: "m.megolm.v1.aes-sha2",
					},
				},
			},
		}
	}

	resp, err := b.lp.GetClient().CreateRoom(req)
	if err != nil {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot create a new room: %v", err)
		return
	}
	roomID := resp.RoomID
	_, err = b.startThread(roomID, userID, hub, false)
	if err != nil {
		// log handled in the startThread
		return
	}
	newEvent := &event.Event{
		Sender: evt.Sender,
		RoomID: roomID,
	}
	newContent := &event.MessageEventContent{
		Body:    b.txt.Start,
		MsgType: event.MsgNotice,
	}
	b.forwardToThread(newEvent, newContent, hub)
}

func (b *Bot) help(evt *event.Event, hub *sentry.Hub) {
	b.log.Debug("help request")
	text := `Honoroit can perform following actions (note that all of them should be sent in a thread:

` + b.prefix + ` done - close the current request. Customer will receive a message about that and bot will leave the customer's room, thead topic will be prefixed with "[DONE]" suffixed with timestamp

` + b.prefix + ` rename TEXT - replaces thread topic text to the TEXT

` + b.prefix + ` invite - invite yourself into the customer's room
`
	content := event.MessageEventContent{
		MsgType:   event.MsgNotice,
		Body:      text,
		RelatesTo: b.getRelatesTo(evt),
	}
	_, err := b.lp.Send(b.roomID, &content)
	if err != nil {
		b.Error(b.roomID, b.getRelatesTo(evt), hub, "cannot send help message: %v", err)
	}
}

func (b *Bot) getRelatesTo(evt *event.Event) *event.RelatesTo {
	content := evt.Content.AsMessage()
	if content == nil {
		return nil
	}

	return content.RelatesTo
}
