package matrix

import (
	"strings"
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
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

func (b *Bot) runCommand(command string, evt *event.Event) {
	switch command {
	case "done", "complete", "close":
		go metrics.RequestDone()
		b.closeRequest(evt)
	case "rename":
		b.renameRequest(evt)
	case "invite":
		b.inviteRequest(evt)
	case "start":
		b.startRequest(evt)
		go metrics.RequestNew()
	case "count":
		b.countRequest(evt)
	case "config":
		b.handleConfig(evt)
	case "note":
		// do nothing
		return
	default:
		b.help(evt)
	}
}

func (b *Bot) renameRequest(evt *event.Event) {
	b.log.Debug().Msg("renaming a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Notice(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I rename your request.", b.getRelatesTo(evt))
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
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

	err = b.replace(threadID, "", "", command, commandFormatted)
	if err != nil {
		b.Notice(b.roomID, err.Error(), b.getRelatesTo(evt))
	}
}

func (b *Bot) closeRequest(evt *event.Event) {
	b.log.Debug().Msg("closing a request")
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Notice(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I close your request.", b.getRelatesTo(evt))
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
		return
	}

	threadEvt, err := b.lp.GetClient().GetEvent(b.roomID, threadID)
	if err != nil {
		b.Notice(b.roomID, err.Error(), b.getRelatesTo(evt))
		return
	}

	roomID, err := b.findRoomID(threadID)
	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
		return
	}

	b.Notice(evt.RoomID, b.cfg.Get(config.TextDone.Key))

	var oldbody string
	if threadEvt.Content.AsMessage() != nil {
		oldbody = strings.Replace(threadEvt.Content.AsMessage().Body, b.cfg.Get(config.TextPrefixOpen.Key), "", 1)
	}
	timestamp := time.Now().UTC().Format("2006/01/02 15:04:05 MST")
	err = b.replace(threadID, b.cfg.Get(config.TextPrefixDone.Key)+" ", oldbody+" ("+timestamp+")", "", "")
	if err != nil {
		b.Notice(b.roomID, err.Error(), b.getRelatesTo(evt))
	}

	_, err = b.lp.GetClient().LeaveRoom(roomID)
	if err != nil {
		// do not send a message when already left
		if !strings.Contains(err.Error(), "M_FORBIDDEN") {
			b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
		}
	}
	b.removeMapping(threadID.String())
	b.removeMapping(roomID.String())
}

func (b *Bot) inviteRequest(evt *event.Event) {
	content := evt.Content.AsMessage()
	relation := content.RelatesTo
	if relation == nil {
		b.Notice(evt.RoomID, "the message doesn't relate to any thread, so I don't know how can I invite you.", b.getRelatesTo(evt))
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
	_, err = b.lp.GetClient().InviteUser(roomID, &mautrix.ReqInviteUser{
		Reason: "you've asked for that",
		UserID: evt.Sender,
	})

	if err != nil {
		b.Notice(evt.RoomID, err.Error(), b.getRelatesTo(evt))
	}
}

func (b *Bot) startRequest(evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) < 2 {
		b.Notice(b.roomID, "cannot start a new matrix room - MXID is not specified", b.getRelatesTo(evt))
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
		b.Notice(b.roomID, err.Error(), b.getRelatesTo(evt))
		return
	}
	roomID := resp.RoomID
	_, err = b.startThread(roomID, userID, false)
	if err != nil {
		// log handled in the startThread
		return
	}
	newEvent := &event.Event{
		Sender: evt.Sender,
		RoomID: roomID,
	}
	newContent := &event.MessageEventContent{
		Body:    b.cfg.Get(config.TextStart.Key),
		MsgType: event.MsgNotice,
	}
	b.forwardToThread(newEvent, newContent)
}

func (b *Bot) countRequest(evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) < 2 {
		b.Notice(b.roomID, "cannot count a request - MXID is not specified", b.getRelatesTo(evt))
		return
	}
	userID := id.UserID(command[1])

	eventID, err := b.newThread(b.cfg.Get(config.TextPrefixDone.Key), userID)
	if err != nil {
		return
	}

	fullContent := &event.Content{
		Parsed: &event.MessageEventContent{
			Body:    b.cfg.Get(config.TextCount.Key),
			MsgType: event.MsgNotice,
			RelatesTo: &event.RelatesTo{
				Type:    ThreadRelation,
				EventID: eventID,
			},
		},
		Raw: map[string]interface{}{
			"event_id": evt.ID,
		},
	}
	_, err = b.lp.Send(b.roomID, fullContent)
	if err != nil {
		b.Notice(b.roomID, err.Error())
	}
}

func (b *Bot) handleConfig(evt *event.Event) {
	command := b.parseCommand(evt.Content.AsMessage().Body)
	if len(command) == 0 {
		return
	}

	switch len(command) {
	case 1:
		b.listConfigOptions()
	case 2:
		b.listConfigOption(command[1])
	default:
		b.setConfigOption(command[1], strings.Join(command[2:], " "))
	}
}

func (b *Bot) listConfigOptions() {
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
	content := format.RenderMarkdown(txt.String(), true, true)
	content.MsgType = event.MsgNotice
	_, err := b.lp.Send(b.roomID, &content)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send config options list")
	}
}

func (b *Bot) listConfigOption(key string) {
	key = strings.ToLower(key)
	option := config.Options.Find(key)
	if option == nil {
		b.Notice(b.roomID, "no such option")
		return
	}

	value := b.cfg.Get(key)
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

	content := format.RenderMarkdown(txt.String(), true, true)
	content.MsgType = event.MsgNotice
	_, err := b.lp.Send(b.roomID, &content)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send config option info")
	}
}

func (b *Bot) setConfigOption(key, value string) {
	option := config.Options.Find(strings.ToLower(key))
	if option == nil {
		b.Notice(b.roomID, "no such option")
		return
	}
	b.cfg.Set(option.Key, option.Sanitizer(value)).Save()

	b.Notice(b.roomID, key+" has been updated, new value: `"+value+"`")
}

func (b *Bot) help(evt *event.Event) {
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
	content := event.MessageEventContent{
		MsgType:   event.MsgNotice,
		Body:      text,
		RelatesTo: b.getRelatesTo(evt),
	}
	_, err := b.lp.Send(b.roomID, &content)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send help message")
	}
}

func (b *Bot) getRelatesTo(evt *event.Event) *event.RelatesTo {
	content := evt.Content.AsMessage()
	if content == nil {
		return nil
	}

	return content.RelatesTo
}
