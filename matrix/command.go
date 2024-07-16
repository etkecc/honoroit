package matrix

import (
	"context"
	"strings"

	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
	"gitlab.com/etke.cc/honoroit/metrics"
)

func (b *Bot) parseCommand(message string) []string {
	if message == "" {
		return nil
	}

	if !strings.HasPrefix(message, b.prefix) {
		return nil
	}

	message = strings.Replace(message, b.prefix, "", 1)
	b.log.Debug().Str("command", message).Msg("received a command")
	return strings.Split(strings.TrimSpace(message), " ")
}

func (b *Bot) readCommand(message string) string {
	command := b.parseCommand(strings.TrimSpace(message))
	if len(command) > 0 {
		return command[0]
	}
	return ""
}

func (b *Bot) runCommand(ctx context.Context, command string, evt *event.Event) {
	switch command {
	case "done", "complete", "close":
		go metrics.RequestDone()
		b.closeRequest(ctx, evt, false)
	case "rename":
		b.renameRequest(ctx, evt)
	case "invite":
		b.inviteRequest(ctx, evt)
	case "start":
		b.startRequest(ctx, evt)
		go metrics.RequestNew()
	case "count":
		b.countRequest(ctx, evt)
	case "config":
		b.handleConfig(ctx, evt)
	case "note":
		// do nothing
		return
	default:
		b.help(ctx, evt)
	}
}

func (b *Bot) renameRequest(ctx context.Context, evt *event.Event) {
	b.log.Debug().Msg("renaming a request")
	content := evt.Content.AsMessage()
	relatesTo := linkpearl.EventRelatesTo(evt)
	relation := content.RelatesTo
	if relation == nil {
		b.SendNotice(ctx, evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I rename your request.", nil, relatesTo)
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
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

	err = b.replace(ctx, threadID, "", "", command, commandFormatted)
	if err != nil {
		b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
	}
}

func (b *Bot) closeRequest(ctx context.Context, evt *event.Event, auto bool) {
	b.log.Debug().Msg("closing a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	relatesTo := linkpearl.EventRelatesTo(evt)
	if relation == nil {
		b.SendNotice(ctx, evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I close your request.", nil, relatesTo)
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	threadEvt, err := b.lp.GetClient().GetEvent(ctx, b.roomID, threadID)
	if err != nil {
		b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	roomID, err := b.findRoomID(ctx, threadID)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	var text string
	if auto {
		text = b.cfg.Get(ctx, config.TextDoneAuto.Key)
	} else {
		text = b.cfg.Get(ctx, config.TextDone.Key)
	}
	b.SendNotice(ctx, roomID, text, nil)
	go b.closeIssue(ctx, roomID, threadID, text)

	var oldbody string
	if threadEvt.Content.AsMessage() != nil {
		oldbody = strings.Replace(threadEvt.Content.AsMessage().Body, b.cfg.Get(ctx, config.TextPrefixOpen.Key), "", 1)
	}
	err = b.replace(ctx, threadID, b.cfg.Get(ctx, config.TextPrefixDone.Key)+" ", oldbody, "", "")
	if err != nil {
		b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
	}

	_, err = b.lp.GetClient().LeaveRoom(ctx, roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(linkpearl.UnwrapError(err).Error(), "M_FORBIDDEN") {
			b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		}
	}

	b.removeMapping(ctx, threadID.String())
	b.removeMapping(ctx, roomID.String())
}

func (b *Bot) inviteRequest(ctx context.Context, evt *event.Event) {
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	relatesTo := linkpearl.EventRelatesTo(evt)
	if relation == nil {
		b.SendNotice(ctx, evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I invite you.", nil, relatesTo)
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	roomID, err := b.findRoomID(ctx, threadID)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}
	_, err = b.lp.GetClient().InviteUser(ctx, roomID, &mautrix.ReqInviteUser{
		Reason: "you've asked for that",
		UserID: evt.Sender,
	})
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
	}
}

func (b *Bot) startRequest(ctx context.Context, evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	relatesTo := linkpearl.EventRelatesTo(evt)
	if len(command) < 2 {
		b.SendNotice(ctx, b.roomID, "cannot start a new matrix room - MXID is not specified", nil, relatesTo)
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

	resp, err := b.lp.GetClient().CreateRoom(ctx, req)
	if err != nil {
		b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}
	roomID := resp.RoomID
	_, err = b.startThread(ctx, roomID, userID, false)
	if err != nil {
		// log handled in the startThread
		return
	}
	newEvent := &event.Event{
		Sender: evt.Sender,
		RoomID: roomID,
	}
	newContent := &event.MessageEventContent{
		Body:    b.cfg.Get(ctx, config.TextStart.Key),
		MsgType: event.MsgNotice,
	}
	b.forwardToThread(ctx, newEvent, newContent)
}

func (b *Bot) countRequest(ctx context.Context, evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) < 2 {
		b.SendNotice(ctx, b.roomID, "cannot count a request - MXID is not specified", nil, linkpearl.EventRelatesTo(evt))
		return
	}
	userID := id.UserID(command[1])

	eventID, _, err := b.newThread(ctx, b.cfg.Get(ctx, config.TextPrefixDone.Key), userID)
	if err != nil {
		return
	}

	b.SendNotice(ctx, b.roomID, b.cfg.Get(ctx, config.TextCount.Key), map[string]any{"event_id": evt.ID}, linkpearl.RelatesTo(eventID))
}

func (b *Bot) handleConfig(ctx context.Context, evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) == 0 {
		return
	}

	switch len(command) {
	case 1:
		b.listConfigOptions(ctx, evt)
	case 2:
		b.listConfigOption(ctx, evt, command[1])
	default:
		b.setConfigOption(ctx, evt, command[1], strings.Join(command[2:], " "))
	}
}

func (b *Bot) listConfigOptions(ctx context.Context, evt *event.Event) {
	var txt strings.Builder
	txt.WriteString("The following config options are available:\n")
	for _, option := range config.Options {
		txt.WriteString("* `")
		txt.WriteString(b.prefix)
		txt.WriteString(" config ")
		txt.WriteString(option.Key)
		txt.WriteString("` VALUE - ")
		txt.WriteString(option.Description)
		txt.WriteString("\n")
	}
	b.SendNotice(ctx, evt.RoomID, txt.String(), nil, linkpearl.EventRelatesTo(evt))
}

func (b *Bot) listConfigOption(ctx context.Context, evt *event.Event, key string) {
	key = strings.ToLower(key)
	option := config.Options.Find(key)
	if option == nil {
		b.SendNotice(ctx, b.roomID, "no such option", nil, linkpearl.EventRelatesTo(evt))
		return
	}

	value := b.cfg.Get(ctx, key)
	if value == "" {
		value = option.Default
	}

	var txt strings.Builder
	txt.WriteString(key)
	txt.WriteString(" - ")
	txt.WriteString(option.Description)

	txt.WriteString("\nCurrent value: `")
	txt.WriteString(value)
	txt.WriteString("`\n")

	txt.WriteString("You can change that option using the following command:\n`")
	txt.WriteString(b.prefix)
	txt.WriteString(" config ")
	txt.WriteString(key)
	txt.WriteString(" NEW VALUE`")

	b.SendNotice(ctx, b.roomID, txt.String(), nil, linkpearl.EventRelatesTo(evt))
}

func (b *Bot) setConfigOption(ctx context.Context, evt *event.Event, key, value string) {
	option := config.Options.Find(strings.ToLower(key))
	if option == nil {
		b.SendNotice(ctx, b.roomID, "no such option", nil, linkpearl.EventRelatesTo(evt))
		return
	}
	b.cfg.Set(option.Key, option.Sanitizer(value)).Save(ctx)

	b.SendNotice(ctx, b.roomID, key+" has been updated, new value: `"+value+"`", nil, linkpearl.EventRelatesTo(evt))
}

func (b *Bot) help(ctx context.Context, evt *event.Event, preamble ...string) {
	text := `Honoroit can perform following actions (note that all of them should be sent in a thread:

` + b.prefix + ` done - close the current request. Customer will receive a message about that and bot will leave the customer's room, thead topic will be prefixed with "[DONE]" suffixed with timestamp

` + b.prefix + ` rename TEXT - replaces thread topic text to the TEXT

` + b.prefix + ` note NOTE - a message prefixed with "!ho note" won't be sent anywhere, it's a safe place to keep notes for other operations in a thread with a customer

` + b.prefix + ` invite - invite yourself into the customer's room

` + b.prefix + ` start MXID - start a conversation with a MXID (like a new thread, but initialized by operator)

` + b.prefix + ` count MXID - count a request from MXID and their homeserver, but don't actually create a room or invite them

` + b.prefix + ` config - list all config options with descriptions

` + b.prefix + ` config KEY - get config KEY's value and description

` + b.prefix + ` config KEY VALUE - set config KEY's value to VALUE
`

	if len(preamble) > 0 {
		text = preamble[0] + "\n\n" + text
	}

	b.SendNotice(ctx, b.roomID, text, nil, linkpearl.EventRelatesTo(evt))
}
