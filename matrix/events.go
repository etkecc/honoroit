package matrix

import (
	"context"
	"slices"

	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) countCustomerRequests(ctx context.Context, userID id.UserID) (int, int, error) {
	var user int
	var hs int
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

func (b *Bot) getName(ctx context.Context, userID id.UserID) string {
	name := userID.String()
	dnresp, err := b.lp.GetClient().GetDisplayName(ctx, userID)
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot get display name")
	}
	if dnresp != nil {
		name = dnresp.DisplayName + " (" + name + ")"
	}

	return name
}

func (b *Bot) getStatus(userID id.UserID) (string, string) {
	hostStatus := ""
	userStatus := ""
	mxids, err := b.psd.GetMXIDs(userID.Homeserver())
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot check psd")
	}
	if len(mxids) > 0 {
		hostStatus = "👥"
	}
	if slices.Contains(mxids, userID.String()) {
		userStatus = "👤"
	}

	return userStatus, hostStatus
}
