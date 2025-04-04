package matrix

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"

	"github.com/etkecc/honoroit/internal/matrix/config"
	"github.com/etkecc/honoroit/internal/metrics"
)

func (b *Bot) greetings(ctx context.Context, userID id.UserID, roomID id.RoomID) {
	ownServer := userID.Homeserver() == b.lp.GetClient().UserID.Homeserver()
	if !ownServer {
		_, requests, err := b.countCustomerRequests(ctx, userID)
		if err != nil {
			b.log.Error().Err(err).Str("userID", userID.String()).Msg("cannot calculate count of the support requests")
		}
		requestsStr := humanize.Ordinal(requests + 1) // including current request
		b.SendNotice(ctx, roomID, fmt.Sprintf(b.cfg.Get(ctx, config.TextGreetingsCustomer.Key), requestsStr), nil)
		return
	}
	b.SendNotice(ctx, roomID, b.cfg.Get(ctx, config.TextGreetings.Key), nil)
}

func (b *Bot) handle(ctx context.Context, evt *event.Event) {
	err := b.lp.GetClient().MarkRead(ctx, evt.RoomID, evt.ID)
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
		b.forwardToThread(ctx, evt, content)
		return
	}

	// message sent from threads room
	// special command
	if command := b.readCommand(content.Body); command != "" {
		b.runCommand(ctx, command, evt)
		return
	}

	// not a command, but a message
	go metrics.MessagesOperator()
	b.forwardToCustomer(ctx, evt, content)
}

func (b *Bot) replace(ctx context.Context, eventID id.EventID, prefix, suffix, body, formattedBody string) error {
	evt, err := b.lp.GetClient().GetEvent(ctx, b.roomID, eventID)
	if err != nil {
		b.log.Error().Err(err).Str("eventID", eventID.String()).Msg("cannot find event to replace")
		b.SendNotice(ctx, b.roomID, "cannot find event to replace", nil)
		return err
	}

	linkpearl.ParseContent(evt, b.log)
	content := evt.Content.AsMessage()
	b.clearPrefix(ctx, content)
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

	_, err = b.lp.Send(ctx, b.roomID, content)
	return err
}

func (b *Bot) clearPrefix(ctx context.Context, content *event.MessageEventContent) {
	for _, prefix := range []string{b.cfg.Get(ctx, config.TextPrefixOpen.Key), b.cfg.Get(ctx, config.TextPrefixDone.Key)} {
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

func (b *Bot) startThread(ctx context.Context, roomID id.RoomID, userID id.UserID, greet bool) (id.EventID, error) {
	mukey := "start_thread_" + roomID.String()
	b.mu.Lock(mukey)
	defer b.mu.Unlock(mukey)
	isSilent := b.cfg.Get(ctx, config.Silent.Key) == "true"

	eventID, err := b.findEventID(ctx, roomID)
	if err != nil && !errors.Is(err, errNotMapped) {
		b.log.Error().Err(err).Str("userID", userID.String()).Str("roomID", roomID.String()).Msg("user tried to send a message from the room, but account data operation failed")
		if !isSilent {
			b.SendNotice(ctx, roomID, b.cfg.Get(ctx, config.TextError.Key), nil)
		}
		return "", err
	}

	if eventID != "" {
		return eventID, nil
	}

	var issueID int64
	eventID, issueID, err = b.newThread(ctx, b.cfg.Get(ctx, config.TextPrefixOpen.Key), userID)
	if err != nil {
		return "", err
	}

	b.setMapping(ctx, roomID.String(), eventID.String())
	b.setMapping(ctx, eventID.String(), roomID.String())
	if issueID != 0 {
		issueIDStr := strconv.Itoa(int(issueID))
		b.setRedmineMapping(ctx, issueIDStr, eventID.String())
		b.setRedmineMapping(ctx, eventID.String(), issueIDStr)
		b.setRedmineMapping(ctx, roomID.String(), issueIDStr)
	}

	if greet && !isSilent {
		b.greetings(ctx, userID, roomID)
	}
	return eventID, nil
}

func (b *Bot) newThread(ctx context.Context, prefix string, userID id.UserID) (id.EventID, int64, error) {
	customerRequests, hsRequests, err := b.countCustomerRequests(ctx, userID)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID.String()).Msg("cannot calculate count of the support requests")
	}
	customerRequestsStr := humanize.Ordinal(customerRequests + 1) // including current request
	hsRequestsStr := humanize.Ordinal(hsRequests + 1)             // including current request
	raw := map[string]any{
		"customer":   userID,
		"homeserver": userID.Homeserver(),
	}

	name, _ := b.getName(ctx, userID)
	eventID := b.SendNotice(ctx, b.roomID, fmt.Sprintf("%s %s request from %s (%s by %s)", prefix, hsRequestsStr, userID.Homeserver(), customerRequestsStr, name), raw)
	if eventID == "" {
		b.SendNotice(ctx, b.roomID, "user "+userID.String()+" tried to send a message, but thread creation failed", nil)
		return "", 0, err
	}

	key := "redmine_" + userID.Homeserver()
	b.mu.Lock(key)
	defer b.mu.Unlock(key)

	threadURL := fmt.Sprintf("https://matrix.to/#/%s/%s", b.roomID, eventID)
	issueID, err := b.redmine.NewIssue(
		fmt.Sprintf("%s request from %s (%s by %s)", hsRequestsStr, userID.Homeserver(), customerRequestsStr, name),
		"Matrix",
		userID.String(),
		fmt.Sprintf("Matrix thread: [%s](%s)", threadURL, threadURL),
	)
	if err != nil {
		b.log.Error().Err(err).Str("userID", userID.String()).Msg("cannot create a new issue in Redmine")
	}
	return eventID, issueID, nil
}

func (b *Bot) forwardToCustomer(ctx context.Context, evt *event.Event, content *event.MessageEventContent) {
	relatesTo := linkpearl.EventRelatesTo(evt)
	if content.RelatesTo == nil {
		if b.cfg.Get(ctx, config.IgnoreNoThread.Key) == "true" {
			return
		}
		b.SendNotice(ctx, evt.RoomID, "the message doesn't relate to any thread, so I don't know where to forward it.", nil, relatesTo)
		return
	}
	threadID, err := b.findThread(evt)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	roomID, err := b.findRoomID(ctx, threadID)
	if err != nil {
		if errors.Is(err, errNotMapped) {
			b.help(ctx, evt, linkpearl.UnwrapError(err).Error())
			return
		}
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
		return
	}

	content.RelatesTo = nil
	b.clearReply(content)
	go b.updateIssue(ctx, true, evt.Sender.String(), threadID, content)
	fullContent := &event.Content{
		Parsed: content,
		Raw: map[string]any{
			"event_id": evt.ID,
		},
	}
	_, err = b.lp.Send(ctx, roomID, fullContent)
	if err != nil {
		b.SendNotice(ctx, evt.RoomID, linkpearl.UnwrapError(err).Error(), nil, relatesTo)
	}
}

func (b *Bot) forwardToThread(ctx context.Context, evt *event.Event, content *event.MessageEventContent) {
	b.mu.Lock(evt.RoomID.String())
	defer b.mu.Unlock(evt.RoomID.String())
	isSilent := b.cfg.Get(ctx, config.Silent.Key) == "true"

	eventID, err := b.startThread(ctx, evt.RoomID, evt.Sender, true)
	if err != nil {
		if !isSilent {
			b.SendNotice(ctx, evt.RoomID, b.cfg.Get(ctx, config.TextError.Key), nil)
		}
		return
	}
	originalContent := *content
	go b.updateIssue(ctx, false, evt.Sender.String(), eventID, &originalContent)

	bodyMD := content.Body
	nameMD, nameHTML := b.getName(ctx, evt.Sender)
	if content.Body != "" {
		content.Body = nameMD + ":\n" + content.Body
	}
	content.Format = event.FormatHTML
	if content.FormattedBody != "" {
		content.FormattedBody = nameHTML + ":<br>" + content.FormattedBody
	} else {
		var formattedBody string
		formatted := format.RenderMarkdown(bodyMD, true, true)
		if formatted.FormattedBody == "" {
			formattedBody = nameHTML + ":<br>" + bodyMD
		} else {
			formattedBody = nameHTML + ":<br>" + formatted.FormattedBody
		}
		content.FormattedBody = formattedBody
	}
	content.RelatesTo = linkpearl.RelatesTo(eventID)

	fullContent := &event.Content{
		Parsed: content,
		Raw: map[string]any{
			"event_id": evt.ID,
		},
	}
	if profile := b.getMSC4144Profie(ctx, evt.Sender); profile != nil {
		fullContent.Raw["com.beeper.per_message_profile"] = profile
	}

	_, err = b.lp.Send(ctx, b.roomID, fullContent)
	if err != nil {
		b.log.Error().Err(err).Str("userID", evt.Sender.String()).Str("roomID", evt.RoomID.String()).Msg("user tried to send a message, but creation of the thread failed")
		if !isSilent {
			b.SendNotice(ctx, evt.RoomID, b.cfg.Get(ctx, config.TextError.Key), nil)
		}
	}
}
