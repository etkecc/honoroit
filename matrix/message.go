package matrix

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
	"gitlab.com/etke.cc/honoroit/metrics"
)

func (b *Bot) Notice(roomID id.RoomID, message string, relatesTo ...*event.RelatesTo) {
	var relates *event.RelatesTo
	if len(relatesTo) > 0 {
		relates = relatesTo[0]
	}

	content := format.RenderMarkdown(message, true, true)
	content.MsgType = event.MsgNotice
	content.RelatesTo = relates
	_, err := b.lp.Send(roomID, &content)
	if err != nil {
		b.log.Error().Err(err).Str("roomID", roomID.String()).Msg("cannot send a notice")
	}
}

func (b *Bot) greetings(roomID id.RoomID) {
	b.Notice(roomID, b.cfg.Get(config.TextGreetings.Key))
}

func (b *Bot) handle(evt *event.Event) {
	err := b.lp.GetClient().MarkRead(evt.RoomID, evt.ID)
	if err != nil {
		b.log.Warn().Err(err).Msg("cannot send mark event")
	}

	content := evt.Content.AsMessage()
	if content == nil {
		b.log.Error().Msg("cannot parse the message")
		return
	}

	// ignore notices
	if content.MsgType == event.MsgNotice {
		return
	}

	// message sent by client
	if evt.RoomID != b.roomID {
		go metrics.MessagesCustomer(evt.Sender)
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
	go metrics.MessagesOperator()
	b.forwardToCustomer(evt, content)
}

func (b *Bot) replace(eventID id.EventID, prefix string, suffix string, body string, formattedBody string) error {
	evt, err := b.lp.GetClient().GetEvent(b.roomID, eventID)
	if err != nil {
		b.log.Error().Err(err).Str("eventID", eventID.String()).Msg("cannot find event to replace")
		b.Notice(b.roomID, "cannot find event to replace")
		return err
	}

	err = evt.Content.ParseRaw(event.EventMessage)
	if err != nil {
		b.log.Error().Err(err).Str("eventID", eventID.String()).Msg("cannot parse thread topic event")
		b.Notice(b.roomID, "cannot parse thread topic event")
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
	content.Body = body
	content.FormattedBody = formattedBody
	content.SetEdit(eventID)

	_, err = b.lp.Send(b.roomID, content)
	return err
}

func (b *Bot) clearPrefix(content *event.MessageEventContent) {
	for _, prefix := range []string{b.cfg.Get(config.TextPrefixOpen.Key), b.cfg.Get(config.TextPrefixDone.Key)} {
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

func (b *Bot) startThread(roomID id.RoomID, userID id.UserID, greet bool) (id.EventID, error) {
	mukey := "start_thread_" + roomID.String()
	b.lock(mukey)
	defer b.unlock(mukey)

	eventID, err := b.findEventID(roomID)
	if err != nil && err != errNotMapped {
		b.log.Error().Err(err).Str("userID", userID.String()).Str("roomID", roomID.String()).Msg("user tried to send a message from the room, but account data operation failed")
		b.Notice(roomID, b.cfg.Get(config.TextError.Key))
		return "", err
	}

	if eventID != "" {
		return eventID, nil
	}

	eventID, err = b.newThread(b.cfg.Get(config.TextPrefixOpen.Key), userID)
	if err != nil {
		return "", err
	}

	b.setMapping(roomID.String(), eventID.String())
	b.setMapping(eventID.String(), roomID.String())
	if greet {
		b.greetings(roomID)
	}
	return eventID, nil
}

func (b *Bot) newThread(prefix string, userID id.UserID) (id.EventID, error) {
	customerRequests, hsRequests, err := b.countCustomerRequests(userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID.String()).Msg("cannot calculate count of the support requests")
	}
	customerRequestsStr := humanize.Ordinal(customerRequests + 1) // including current request
	hsRequestsStr := humanize.Ordinal(hsRequests + 1)             // including current request
	content := &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    fmt.Sprintf("%s %s request from %s (%s by %s)", prefix, hsRequestsStr, userID.Homeserver(), customerRequestsStr, b.getName(userID)),
	}

	fullContent := &event.Content{
		Parsed: content,
		Raw: map[string]interface{}{
			"customer":   userID,
			"homeserver": userID.Homeserver(),
		},
	}

	eventID, err := b.lp.Send(b.roomID, fullContent)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID.String()).Msg("user tried to send a message, but thread creation failed")
		b.Notice(b.roomID, "user "+userID.String()+"tried to send a message, but thread creation failed")
		return "", err
	}
	return eventID, nil
}

func (b *Bot) forwardToCustomer(evt *event.Event, content *event.MessageEventContent) {
	relation := content.RelatesTo
	if relation == nil {
		if b.cfg.Get(config.IgnoreNoThread.Key) == "true" {
			return
		}
		b.Notice(evt.RoomID, "the message doesn't relate to any thread, so I don't know where to forward it.", b.getRelatesTo(evt))
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
		return
	}

	roomID, err := b.findRoomID(threadID)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
		return
	}

	content.RelatesTo = nil
	b.clearReply(content)
	fullContent := &event.Content{
		Parsed: content,
		Raw: map[string]interface{}{
			"event_id": evt.ID,
		},
	}
	_, err = b.lp.Send(roomID, fullContent)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
	}
}

func (b *Bot) forwardToThread(evt *event.Event, content *event.MessageEventContent) {
	b.lock(evt.RoomID.String())
	defer b.unlock(evt.RoomID.String())

	eventID, err := b.startThread(evt.RoomID, evt.Sender, true)
	if err != nil {
		b.Notice(evt.RoomID, b.cfg.Get(config.TextError.Key))
		return
	}

	content.SetRelatesTo(&event.RelatesTo{
		Type:    ThreadRelation,
		EventID: eventID,
	})

	fullContent := &event.Content{
		Parsed: content,
		Raw: map[string]interface{}{
			"event_id": evt.ID,
		},
	}
	_, err = b.lp.Send(b.roomID, fullContent)
	if err != nil {
		b.log.Error().Err(err).Str("userID", evt.Sender.String()).Str("roomID", evt.RoomID.String()).Msg("user tried to send a message, but creation of the thread failed")
		b.Notice(evt.RoomID, b.cfg.Get(config.TextError.Key))
	}
}
