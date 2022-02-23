package matrix

import (
	"github.com/getsentry/sentry-go"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (b *Bot) initSync() {
	b.lp.OnEventType(
		event.StateMember,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onEmpty(evt)
		},
	)
	b.lp.OnEventType(
		event.EventMessage,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onMessage(source, evt)
		},
	)
	b.lp.OnEventType(
		event.EventEncrypted,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onEncryptedMessage(source, evt)
		},
	)
}

func (b *Bot) onEmpty(evt *event.Event) {
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{ID: evt.Sender.String()})
		scope.SetContext("event", map[string]string{
			"id":     evt.ID.String(),
			"room":   evt.RoomID.String(),
			"sender": evt.Sender.String(),
		})
	})

	members := b.lp.GetStore().GetRoomMembers(evt.RoomID)
	if len(members) >= 1 && members[0] != b.lp.GetClient().UserID {
		return
	}
	eventID, err := b.findEventID(evt.RoomID)
	// there is no thread for that room
	if err == errNotMapped {
		return
	}
	if err != nil {
		b.log.Error("cannot find eventID for room %s: %v", evt.RoomID, err)
		return
	}

	content := &event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    b.txt.EmptyRoom,
		RelatesTo: &event.RelatesTo{
			Type:    ThreadRelation,
			EventID: eventID,
		},
	}

	_, err = b.lp.Send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, hub, "cannot send a notice about empty room %s: %v", evt.RoomID, err)
	}
}

func (b *Bot) onMessage(_ mautrix.EventSource, evt *event.Event) {
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

func (b *Bot) onEncryptedMessage(_ mautrix.EventSource, evt *event.Event) {
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
