package matrix

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/etke.cc/go/mxidwc"
	"gitlab.com/etke.cc/linkpearl"
	"golang.org/x/exp/slices"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
)

func (b *Bot) initSync() {
	b.lp.SetJoinPermit(b.joinPermit)
	b.lp.OnEventType(
		event.StateMember,
		func(ctx context.Context, evt *event.Event) {
			go b.onMembership(ctx, evt)
		},
	)
	b.lp.OnEventType(
		event.EventReaction,
		func(ctx context.Context, evt *event.Event) {
			go b.onReaction(ctx, evt)
		},
	)
	b.lp.OnEventType(
		event.EventEncrypted,
		func(ctx context.Context, evt *event.Event) {
			go b.onEncryptedMessage(ctx, evt)
		},
	)
	b.lp.OnEventType(
		event.EventMessage,
		func(ctx context.Context, evt *event.Event) {
			go b.onMessage(ctx, evt)
		},
	)
}

// joinPermit is called by linkpearl when processing "invite" events and deciding if rooms should be auto-joined or not
func (b *Bot) joinPermit(ctx context.Context, evt *event.Event) bool {
	allowed, err := parseMXIDpatterns(b.cfg.Get(ctx, config.AllowedUsers.Key))
	if err != nil {
		b.log.Error().Err(err).Msg("cannot parse MXID patterns")
		return false
	}

	if !mxidwc.Match(evt.Sender.String(), allowed) {
		b.log.Debug().Str("userID", evt.Sender.String()).Msg("Rejecting room invitation from unallowed user")
		return false
	}

	return true
}

func (b *Bot) onJoin(ctx context.Context, evt *event.Event, threadID id.EventID) {
	name, _ := b.getName(ctx, evt.Sender)
	b.SendNotice(ctx, b.roomID, fmt.Sprintf(b.cfg.Get(ctx, config.TextJoin.Key), name), nil, linkpearl.RelatesTo(threadID))
}

func (b *Bot) onInvite(ctx context.Context, evt *event.Event, threadID id.EventID) {
	nameSender, _ := b.getName(ctx, evt.Sender)
	nameTarget, _ := b.getName(ctx, id.UserID(evt.GetStateKey()))
	b.SendNotice(ctx, b.roomID, fmt.Sprintf(b.cfg.Get(ctx, config.TextInvite.Key), nameSender, nameTarget), nil, linkpearl.RelatesTo(threadID))
}

func (b *Bot) onLeave(ctx context.Context, evt *event.Event, threadID id.EventID) {
	name, _ := b.getName(ctx, id.UserID(evt.GetStateKey()))
	b.SendNotice(ctx, b.roomID, fmt.Sprintf(b.cfg.Get(ctx, config.TextLeave.Key), name), nil, linkpearl.RelatesTo(threadID))

	members, err := b.lp.GetClient().StateStore.GetRoomJoinedOrInvitedMembers(ctx, evt.RoomID)
	if err != nil {
		b.log.Error().Err(err).Str("roomID", evt.RoomID.String()).Msg("cannot get joined or invited members")
		return
	}

	count := len(members)
	if count == 1 && members[0] == b.lp.GetClient().UserID {
		b.SendNotice(ctx, b.roomID, b.cfg.Get(ctx, config.TextEmptyRoom.Key), nil, linkpearl.RelatesTo(threadID))
	}
}

func (b *Bot) onMembership(ctx context.Context, evt *event.Event) {
	// ignore own events
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(ctx, config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}
	eventID, err := b.findEventID(ctx, evt.RoomID)
	// there is no thread for that room
	if errors.Is(err, errNotMapped) {
		return
	}
	if err != nil {
		b.log.Error().Err(err).Msg("cannot find eventID for the room")
		return
	}

	switch evt.Content.AsMember().Membership { //nolint:exhaustive // we don't care about other membership types
	case event.MembershipJoin:
		b.onJoin(ctx, evt, eventID)
	case event.MembershipInvite:
		b.onInvite(ctx, evt, eventID)
	case event.MembershipLeave, event.MembershipBan:
		b.onLeave(ctx, evt, eventID)
	}
}

func (b *Bot) onReaction(ctx context.Context, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(ctx, config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}

	b.forwardReaction(ctx, evt)
}

func (b *Bot) onEncryptedMessage(ctx context.Context, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(ctx, config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}

	// if the room already mappend, i.e. it already has own thread, ignore, and handle it in the onMessage
	mapping, err := b.getMapping(ctx, evt.RoomID.String())
	if mapping != "" || err == nil {
		return
	}

	// otherwise, send a notice about potential issues with encrypted messages
	content := format.RenderMarkdown(b.cfg.Get(ctx, config.TextGreetingsBeforeEncryption.Key), true, true)
	content.MsgType = event.MsgNotice
	fullContent := &event.Content{Parsed: &content}
	if _, err := b.lp.GetClient().SendMessageEvent(ctx, evt.RoomID, event.EventMessage, fullContent, mautrix.ReqSendEvent{DontEncrypt: true}); err != nil {
		b.log.Error().Err(err).Str("roomID", evt.RoomID.String()).Msg("cannot send a notice about encrypted messages")
	}
}

func (b *Bot) onMessage(ctx context.Context, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(ctx, config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}

	b.handle(ctx, evt)
}
