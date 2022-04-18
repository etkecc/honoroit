package matrix

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func (b *Bot) initSync() {
	b.lp.OnEventType(
		event.StateMember,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onMembership(evt)
		},
	)
	b.lp.OnEventType(
		event.EventMessage,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onMessage(evt)
		},
	)
	b.lp.OnEventType(
		event.EventEncrypted,
		func(_ mautrix.EventSource, evt *event.Event) {
			go b.onEncryptedMessage(evt)
		},
	)
}

func (b *Bot) onJoin(evt *event.Event, threadID id.EventID, hub *sentry.Hub) {
	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    fmt.Sprintf(b.txt.Join, evt.Sender),
		RelatesTo: &event.RelatesTo{
			Type:    ThreadRelation,
			EventID: threadID,
		},
	}

	_, err := b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "cannot send a notice about joined customer in room %s: %v", evt.RoomID, err)
	}
}

func (b *Bot) onInvite(evt *event.Event, threadID id.EventID, hub *sentry.Hub) {
	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    fmt.Sprintf(b.txt.Invite, evt.Sender, evt.GetStateKey()),
		RelatesTo: &event.RelatesTo{
			Type:    ThreadRelation,
			EventID: threadID,
		},
	}

	_, err := b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "cannot send a notice about joined customer in room %s: %v", evt.RoomID, err)
	}
}

func (b *Bot) onLeave(evt *event.Event, threadID id.EventID, hub *sentry.Hub) {
	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    fmt.Sprintf(b.txt.Leave, evt.GetStateKey()),
		RelatesTo: &event.RelatesTo{
			Type:    ThreadRelation,
			EventID: threadID,
		},
	}

	_, err := b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "cannot send a notice about joined customer in room %s: %v", evt.RoomID, err)
	}

	members := b.lp.GetStore().GetRoomMembers(evt.RoomID)
	count := len(members)
	if count == 1 && members[0] == b.lp.GetClient().UserID {
		content := &event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    b.txt.EmptyRoom,
			RelatesTo: &event.RelatesTo{
				Type:    ThreadRelation,
				EventID: threadID,
			},
		}

		_, err = b.lp.Send(b.roomID, content)
		if err != nil {
			b.Error(b.roomID, hub, "cannot send a notice about empty room %s: %v", evt.RoomID, err)
		}
	}
}

func (b *Bot) onMembership(evt *event.Event) {
	// ignore own events
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: evt.Sender.String()})
		scope.SetContext("event", map[string]string{
			"id":     evt.ID.String(),
			"room":   evt.RoomID.String(),
			"sender": evt.Sender.String(),
		})
	})

	eventID, err := b.findEventID(evt.RoomID)
	// there is no thread for that room
	if err == errNotMapped {
		return
	}
	if err != nil {
		b.log.Error("cannot find eventID for room %s: %v", evt.RoomID, err)
		return
	}

	switch evt.Content.AsMember().Membership {
	case event.MembershipJoin:
		b.onJoin(evt, eventID, hub)
	case event.MembershipInvite:
		b.onInvite(evt, eventID, hub)
	case event.MembershipLeave, event.MembershipBan:
		b.onLeave(evt, eventID, hub)
	}
}

func (b *Bot) onMessage(evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: evt.Sender.String()})
		scope.SetContext("event", map[string]string{
			"id":     evt.ID.String(),
			"room":   evt.RoomID.String(),
			"sender": evt.Sender.String(),
		})
	})

	b.handle(evt, hub)
}

func (b *Bot) onEncryptedMessage(evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.lp.GetClient().UserID {
		return
	}

	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: evt.Sender.String()})
		scope.SetContext("event", map[string]string{
			"id":     evt.ID.String(),
			"room":   evt.RoomID.String(),
			"sender": evt.Sender.String(),
		})
	})

	decrypted, err := b.lp.GetMachine().DecryptMegolmEvent(evt)
	if err != nil {
		b.Error(b.roomID, hub, "cannot decrypt a message by %s in %s: %v", evt.Sender, evt.RoomID, err)
		b.Error(evt.RoomID, hub, b.txt.Error)
		return
	}

	b.handle(decrypted, hub)
}
