package matrix

import (
	"context"
	"fmt"
	"strings"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/etkecc/honoroit/internal/matrix/config"
)

// MSC4144Profile represents the profile of the user according to MSC4144
// ref: https://github.com/beeper/matrix-spec-proposals/blob/per-message-profile/proposals/4144-per-message-profile.md
type MSC4144Profile struct {
	ID          string        `json:"id"`          // The id field is required and is an opaque string. Clients may use it to group messages with the same ID like they would group messages from the same sender. For example, bridges would likely set it to the immutable remote user ID.
	DisplayName string        `json:"displayname"` // The displayname field represents the human-readable display name of the user who sent the message. It is recommended that clients use this display name when showing messages to end users.
	AvatarURL   id.ContentURI `json:"avatar_url"`  // The avatar_url field represents the URL of the avatar image that should be used when rendering the message. This URL must be an MXC URI.
}

func (b *Bot) countCustomerRequests(ctx context.Context, userID id.UserID) (user, hs int, err error) {
	var from string
	for {
		resp, err := b.lp.Threads(ctx, b.roomID)
		if err != nil {
			b.log.Error().Err(err).Str("from", from).Str("roomID", b.roomID.String()).Msg("cannot request threads for the room")
			return user, hs, err
		}
		for _, evt := range resp.Chunk {
			if linkpearl.EventContains(evt, "customer", userID.String()) {
				user++
			}
			if linkpearl.EventContains(evt, "homeserver", userID.Homeserver()) {
				hs++
			}
		}
		from = resp.NextBatch
		if resp.NextBatch == "" {
			break
		}
	}

	return user, hs, nil
}

func (b *Bot) getName(ctx context.Context, userID id.UserID) (md, html string) {
	if parts, ok := b.namesCache.Get(userID); ok {
		return parts[0], parts[1]
	}

	md = userID.String()
	html = fmt.Sprintf("<a href=\"https://matrix.to/#/%s\">%s</a>", userID.String(), userID.String())

	dnresp, err := b.lp.GetClient().GetDisplayName(ctx, userID)
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot get display name")
	}
	if dnresp != nil {
		md = fmt.Sprintf("%s (%s)", dnresp.DisplayName, userID.String())
		html = fmt.Sprintf("<a href=\"https://matrix.to/#/%s\">%s</a>", userID.String(), dnresp.DisplayName)
	}

	b.namesCache.Add(userID, [2]string{md, html})
	return md, html
}

// getContentBody returns the body and formatted body of the message,
// taking into account the new content if it exists
func (b *Bot) getContentBody(content *event.MessageEventContent) (body, formattedBody string) {
	if content == nil {
		return "", ""
	}
	body = content.Body
	formattedBody = content.FormattedBody
	if content.NewContent != nil {
		body = content.NewContent.Body
		formattedBody = content.NewContent.FormattedBody
	}
	if formattedBody == "" {
		formattedBody = body
	}
	return body, formattedBody
}

// getMSC4144Profie returns the MSC4144 profile of the user
func (b *Bot) getMSC4144Profie(ctx context.Context, userID id.UserID) *MSC4144Profile {
	if profile, ok := b.profilesCache.Get(userID); ok {
		return profile
	}

	remoteProfile, err := b.lp.GetClient().GetProfile(ctx, userID)
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot get profile")
		return nil
	}
	profile := &MSC4144Profile{
		ID:          userID.String(),
		DisplayName: remoteProfile.DisplayName,
		AvatarURL:   remoteProfile.AvatarURL,
	}
	b.profilesCache.Add(userID, profile)
	return profile
}

// getThreadIDs returns all thread IDs in the operator room
func (b *Bot) getThreadIDs(ctx context.Context, fromToken ...string) []id.EventID {
	threadIDs := []id.EventID{}
	threads, err := b.lp.Threads(ctx, b.roomID, fromToken...)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot get room threads")
		return threadIDs
	}
	for _, thread := range threads.Chunk {
		threadIDs = append(threadIDs, thread.ID)
	}
	if threads.NextBatch == "" {
		return threadIDs
	}

	return append(threadIDs, b.getThreadIDs(ctx, threads.NextBatch)...)
}

// isLastThreadMessageEligible checks if the last message in the thread is eligible for processing
func (b *Bot) isLastThreadMessageEligible(ctx context.Context, evt *event.Event) (eligible, shouldContinue bool) {
	// not a message - ignore
	if evt.Type != event.EventMessage {
		return false, true
	}
	linkpearl.ParseContent(evt, b.log)
	content := evt.Content.AsMessage()
	// not a message - ignore
	if content == nil {
		return false, true
	}
	// not a text - ignore
	if content.MsgType != event.MsgText {
		return false, true
	}
	// thread already closed - nothing to do
	if strings.TrimSpace(content.Body) == b.prefix+" done" {
		return false, false
	}
	// thread was created just to count requests - nothing to do
	if strings.TrimSpace(content.Body) == b.cfg.Get(ctx, config.TextCount.Key) {
		return false, false
	}
	return true, false
}

// getLastThreadMessage returns the last message in the thread
func (b *Bot) getLastThreadMessage(ctx context.Context, threadID id.EventID, fromToken ...string) *event.Event {
	evts, err := b.lp.Relations(ctx, b.roomID, threadID, "m.thread", fromToken...)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot get thread messages")
		return nil
	}
	var lastEvt *event.Event
	for _, evt := range evts.Chunk {
		eligible, shouldContinue := b.isLastThreadMessageEligible(ctx, evt)
		if shouldContinue {
			continue
		}
		if !eligible {
			return nil
		}

		if lastEvt == nil || evt.Timestamp > lastEvt.Timestamp {
			lastEvt = evt
		}
	}

	if evts.NextBatch == "" {
		return lastEvt
	}
	return b.getLastThreadMessage(ctx, threadID, evts.NextBatch)
}

// getLastEdit returns the last edit of the message
func (b *Bot) getLastEdit(ctx context.Context, roomID id.RoomID, eventID id.EventID) *event.Event {
	edits, err := b.lp.Relations(ctx, roomID, eventID, string(event.RelReplace))
	if err != nil {
		b.log.Error().Err(err).Msg("cannot get edits")
		return nil
	}
	if len(edits.Chunk) == 0 {
		return nil
	}

	evt := edits.Chunk[len(edits.Chunk)-1]
	isEncrypted := evt.Type == event.EventEncrypted
	linkpearl.ParseContent(evt, b.log)
	log := b.log.With().
		Str("originalEventID", eventID.String()).
		Str("lastEditID", evt.ID.String()).
		Bool("encrypted", isEncrypted).
		Logger()
	log.Debug().Msg("found last edit")
	if !isEncrypted {
		return evt
	}
	decrypted, derr := b.lp.GetClient().Crypto.Decrypt(ctx, evt)
	if derr != nil {
		log.Error().Err(derr).Msg("cannot decrypt last edit")
		return nil
	}
	log.Debug().Msg("successfully decrypted last edit")
	linkpearl.ParseContent(decrypted, b.log)
	return decrypted
}
