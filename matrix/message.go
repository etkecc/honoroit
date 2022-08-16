package matrix

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/getsentry/sentry-go"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// Error message to the log and matrix room
func (b *Bot) Error(roomID id.RoomID, hub *sentry.Hub, message string, args ...interface{}) {
	b.log.Error(message, args...)
	if hub != nil {
		hub.CaptureException(fmt.Errorf(message, args...))
	} else {
		sentry.CaptureException(fmt.Errorf(message, args...))
	}

	if b.lp == nil {
		return
	}
	// nolint // if something goes wrong here nobody can help...
	b.lp.Send(roomID, &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    "ERROR: " + fmt.Sprintf(message, args...),
	})
}

func (b *Bot) greetings(roomID id.RoomID, userID id.UserID, hub *sentry.Hub) {
	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.Greetings,
	}
	if _, err := b.lp.Send(roomID, content); err != nil {
		b.Error(b.roomID, hub, "cannot send a greetings message to the user %s in the room %s: %v", userID, roomID, err)
	}
}

func (b *Bot) handle(evt *event.Event, hub *sentry.Hub) {
	err := b.lp.GetClient().MarkRead(evt.RoomID, evt.ID)
	if err != nil {
		b.Error(b.roomID, hub, "cannot send mark event to the room %s: %v", evt.RoomID, err)
	}

	content := evt.Content.AsMessage()
	if unsafe.Sizeof(content) == 0 {
		b.Error(evt.RoomID, hub, "cannot parse the message")
		return
	}

	// message sent by client
	if evt.RoomID != b.roomID {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("to", "thread")
		})
		b.forwardToThread(evt, content, hub)
		return
	}

	// message sent from threads room
	// special command
	if command := b.readCommand(content.Body); command != "" {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("to", "command")
		})
		b.runCommand(command, evt, hub)
		return
	}

	// not a command, but a message
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("to", "customer")
	})
	b.forwardToCustomer(evt, content, hub)
}

func (b *Bot) replace(eventID id.EventID, hub *sentry.Hub, prefix string, suffix string, body string, formattedBody string) error {
	evt, err := b.lp.GetClient().GetEvent(b.roomID, eventID)
	if err != nil {
		b.Error(b.roomID, hub, "cannot find event %s: %v", eventID, err)
		return err
	}

	err = evt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.Error(b.roomID, hub, "cannot parse thread topic event %s: %v", eventID, err)
		return err
	}
	content := evt.Content.AsMessage()
	b.clearPrefix(content)
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
	_, err = b.lp.Send(b.roomID, content)
	return err
}

func (b *Bot) clearPrefix(content *event.MessageEventContent) {
	for _, prefix := range b.prefixes {
		index := strings.Index(content.Body, prefix)
		formattedIndex := strings.Index(content.FormattedBody, prefix)
		if index > -1 {
			content.Body = strings.Replace(content.Body, prefix, "", 1)
		}
		if formattedIndex > -1 {
			content.FormattedBody = strings.Replace(content.FormattedBody, prefix, "", 1)
		}
	}
}

// clearReply removes quotation of previous message in reply message, because it may contain sensitive info
func (b *Bot) clearReply(content *event.MessageEventContent) {
	index := strings.Index(content.Body, "> <@")
	formattedIndex := strings.Index(content.FormattedBody, "</mx-reply>")
	if index >= 0 {
		index = strings.Index(content.Body, "\n\n")
		// 2 is length of "\n\n"
		content.Body = content.Body[index+2:]
	}

	if formattedIndex >= 0 {
		// 11 is length of "</mx-reply>"
		content.FormattedBody = content.FormattedBody[formattedIndex+11:]
	}
}

func (b *Bot) startThread(roomID id.RoomID, userID id.UserID, hub *sentry.Hub, greet bool) (id.EventID, error) {
	b.log.Debug("starting new thread for %s request from %s", userID, roomID)
	mukey := "start_thread_" + roomID.String()
	b.lock(mukey)
	defer b.unlock(mukey)

	eventID, err := b.findEventID(roomID)
	if err != nil && err != errNotMapped {
		b.Error(b.roomID, hub, "user %s tried to send a message from room %s, but account data operation returned a error: %v", userID, roomID, err)
		b.Error(roomID, hub, b.txt.Error)
		return "", err
	}

	if eventID != "" {
		b.log.Debug("thread for %s request from %s exists, returning eventID %s instead", userID, roomID, eventID)
		return eventID, nil
	}

	content := &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    b.txt.PrefixOpen + " Request by " + b.getName(userID, hub) + " in " + roomID.String(),
	}

	eventID, err = b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "user %s tried to send a message from room %s, but creation of a thread failed: %v", userID, roomID, err)
		return "", err
	}

	b.saveMapping(roomID, "", eventID)
	if greet {
		b.greetings(roomID, userID, hub)
	}
	return eventID, nil
}

func (b *Bot) forwardToCustomer(evt *event.Event, content *event.MessageEventContent, hub *sentry.Hub) {
	b.log.Debug("forwarding the message to a customer room")
	relation := content.RelatesTo
	if relation == nil {
		if b.ignoreNoThread {
			return
		}
		b.Error(evt.RoomID, hub, "the message doesn't relate to any thread, so I don't know where to forward it.")
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Error(evt.RoomID, hub, "cannot find a thread: %v", err)
		return
	}

	roomID, err := b.findRoomID(threadID)
	if err != nil {
		b.Error(evt.RoomID, hub, err.Error())
		return
	}

	content.RelatesTo = nil
	b.clearReply(content)
	_, err = b.lp.Send(roomID, content)
	if err != nil {
		b.Error(evt.RoomID, hub, err.Error())
	}
}

func (b *Bot) forwardToThread(evt *event.Event, content *event.MessageEventContent, hub *sentry.Hub) {
	b.log.Debug("forwaring a message from customer to the threads rooms")
	eventID, err := b.startThread(evt.RoomID, evt.Sender, hub, true)
	if err != nil {
		b.Error(evt.RoomID, hub, b.txt.Error)
		return
	}

	content.RelatesTo = &event.RelatesTo{
		Type:    ThreadRelation,
		EventID: eventID,
	}

	_, err = b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "user %s tried to send a message from room %s, but creation of a thread failed: %v", evt.Sender, evt.RoomID, err)
		b.Error(evt.RoomID, hub, b.txt.Error)
	}
}

func (b *Bot) getName(userID id.UserID, hub *sentry.Hub) string {
	name := userID.String()
	dnresp, err := b.lp.GetClient().GetDisplayName(userID)
	if err != nil {
		b.Error(b.roomID, hub, "cannot get user %s display name: %v", userID, err)
	}
	if dnresp != nil {
		name = dnresp.DisplayName + " (" + name + ")"
	}

	return name
}
