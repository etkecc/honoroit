package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (b *Bot) initSync() {
	if b.olm != nil {
		b.api.Syncer.(*mautrix.DefaultSyncer).OnSync(b.olm.ProcessSyncResponse)
	}

	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.StateEncryption,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onEncryption(source, evt)
		},
	)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.StateMember,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onMembership(source, evt)
		},
	)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.EventMessage,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onMessage(source, evt)
		},
	)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.EventEncrypted,
		func(source mautrix.EventSource, evt *event.Event) {
			go b.onEncryptedMessage(source, evt)
		},
	)
}

func (b *Bot) onMembership(_ mautrix.EventSource, evt *event.Event) {
	if b.olm != nil {
		b.olm.HandleMemberEvent(evt)
	}
	b.store.SetMembership(evt)

	// autoaccept invites
	b.onInvite(evt)
	// autoleave empty rooms
	b.onEmpty(evt)
}

func (b *Bot) onInvite(evt *event.Event) {
	userID := b.api.UserID.String()
	invite := evt.Content.AsMember().Membership == event.MembershipInvite
	if invite && evt.GetStateKey() == userID {
		_, err := b.api.JoinRoomByID(evt.RoomID)
		if err != nil {
			b.log.Error("cannot join the room %s: %v", evt.RoomID, err)
		}
	}
}

func (b *Bot) onEmpty(evt *event.Event) {
	members := b.store.GetRoomMembers(evt.RoomID)
	if len(members) >= 1 && members[0] != b.api.UserID {
		return
	}

	_, err := b.api.LeaveRoom(evt.RoomID)
	if err != nil {
		b.log.Error("cannot leave room: %v", err)
	}
	b.log.Debug("left room %s because it's empty", evt.RoomID)
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

	_, err = b.send(b.roomID, content)
	if err != nil {
		b.Error(b.roomID, "cannot send a notice about empty room %s: %v", evt.RoomID, err)
	}
}

func (b *Bot) onEncryption(_ mautrix.EventSource, evt *event.Event) {
	b.store.SetEncryptionEvent(evt)
}

func (b *Bot) onMessage(_ mautrix.EventSource, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.api.UserID {
		return
	}

	b.handle(evt)
}

func (b *Bot) onEncryptedMessage(_ mautrix.EventSource, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.api.UserID {
		return
	}

	if b.olm == nil {
		_, err := b.send(evt.RoomID, &event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    "Unfortunately, encrypted rooms is not supported yet. Please, send an unencrypted message",
		})
		if err != nil {
			b.Error(b.roomID, "cannot send a message to an encrypted room: %v", err)
		}
		return
	}

	decrypted, err := b.olm.DecryptMegolmEvent(evt)
	if err != nil {
		b.Error(b.roomID, "cannot decrypt a message by %s in %s: %v", evt.Sender, evt.RoomID, err)
		b.Error(evt.RoomID, b.txt.Error)
		return
	}

	b.handle(decrypted)
}
