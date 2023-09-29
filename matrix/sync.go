package matrix

import (
	"fmt"
	"strings"

	"gitlab.com/etke.cc/go/mxidwc"
	"gitlab.com/etke.cc/linkpearl"
	"golang.org/x/exp/slices"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
)

func (b *Bot) initSync() {
	b.lp.SetJoinPermit(b.joinPermit)
	b.lp.OnEventType(
		event.StateMember,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onMembership(evt)
		},
	)
	b.lp.OnEventType(
		event.EventReaction,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onReaction(evt)
		},
	)
	b.lp.OnEventType(
		event.EventMessage,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onMessage(evt)
		},
	)
}

// joinPermit is called by linkpearl when processing "invite" events and deciding if rooms should be auto-joined or not
func (b *Bot) joinPermit(evt *event.Event) bool {
	allowed, err := parseMXIDpatterns(b.cfg.Get(config.AllowedUsers.Key))
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

func (b *Bot) onJoin(evt *event.Event, threadID id.EventID) {
	b.SendNotice(b.roomID, fmt.Sprintf(b.cfg.Get(config.TextJoin.Key), b.getName(evt.Sender)), nil, linkpearl.RelatesTo(threadID))
}

func (b *Bot) onInvite(evt *event.Event, threadID id.EventID) {
	b.SendNotice(b.roomID, fmt.Sprintf(b.cfg.Get(config.TextInvite.Key), b.getName(evt.Sender), b.getName(id.UserID(evt.GetStateKey()))), nil, linkpearl.RelatesTo(threadID))
}

func (b *Bot) onLeave(evt *event.Event, threadID id.EventID) {
	b.SendNotice(b.roomID, fmt.Sprintf(b.cfg.Get(config.TextLeave.Key), b.getName(id.UserID(evt.GetStateKey()))), nil, linkpearl.RelatesTo(threadID))

	members, err := b.lp.GetClient().StateStore.GetRoomJoinedOrInvitedMembers(evt.RoomID)
	if err != nil {
		b.log.Error().Err(err).Str("roomID", evt.RoomID.String()).Msg("cannot get joined or invited members")
		return
	}

	count := len(members)
	if count == 1 && members[0] == b.lp.GetClient().UserID {
		b.SendNotice(b.roomID, b.cfg.Get(config.TextEmptyRoom.Key), nil, linkpearl.RelatesTo(threadID))
	}
}

func (b *Bot) onMembership(evt *event.Event) {
	// ignore own events
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}
	eventID, err := b.findEventID(evt.RoomID)
	// there is no thread for that room
	if err == errNotMapped {
		return
	}
	if err != nil {
		b.log.Error().Err(err).Msg("cannot find eventID for the room")
		return
	}

	switch evt.Content.AsMember().Membership {
	case event.MembershipJoin:
		b.onJoin(evt, eventID)
	case event.MembershipInvite:
		b.onInvite(evt, eventID)
	case event.MembershipLeave, event.MembershipBan:
		b.onLeave(evt, eventID)
	}
}

func (b *Bot) onReaction(evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}

	b.forwardReaction(evt)
}

func (b *Bot) onMessage(evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	// mautrix 0.15.x migration
	if b.ignoreBefore >= evt.Timestamp {
		return
	}

	// ignore any events in ignored rooms
	if slices.Contains(strings.Split(b.cfg.Get(config.IgnoredRooms.Key), ","), evt.RoomID.String()) {
		return
	}

	b.handle(evt)
}
