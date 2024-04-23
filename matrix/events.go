package matrix

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
)

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

func (b *Bot) getStatus(userID id.UserID) (userStatus, hostStatus string) {
	hostTargets, err := b.psdc.Get(userID.Homeserver())
	if err != nil {
		b.log.Warn().Err(err).Str("host", userID.Homeserver()).Msg("cannot check psd")
	}
	userTargets, err := b.psdc.Get(userID.String())
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot check psd")
	}
	if len(hostTargets) > 0 {
		hostStatus = "👥"
	}
	if len(userTargets) > 0 {
		userStatus = "👤"
	}

	return userStatus, hostStatus
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
