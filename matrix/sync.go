package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (b *Bot) onMembership(_ mautrix.EventSource, evt *event.Event) {
	b.olm.HandleMemberEvent(evt)
	b.store.SetMembership(evt)

	// autoaccept invites
	b.onInvite(evt)
	// autoleave empty rooms
	b.onEmpty(evt)
}

func (b *Bot) onInvite(evt *event.Event) {
	userID := b.userID.String()
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
	if len(members) == 1 && members[0] == b.userID {
		_, err := b.api.LeaveRoom(evt.RoomID)
		if err != nil {
			b.log.Error("cannot leave room: %v", err)
		}
	}
}

func (b *Bot) onEncryption(_ mautrix.EventSource, evt *event.Event) {
	b.store.SetEncryptionEvent(evt)
}

func (b *Bot) onMessage(_ mautrix.EventSource, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.userID {
		return
	}

	b.handle(evt)
}

func (b *Bot) onEncryptedMessage(_ mautrix.EventSource, evt *event.Event) {
	// ignore own messages
	if evt.Sender == b.userID {
		return
	}

	_, err := b.api.SendMessageEvent(evt.RoomID, event.EventMessage, &event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    "Unfortunately, I don't work in encrypted rooms yet. Please, send me an unencrypted message",
	})
	if err != nil {
		b.Error(b.roomID, "cannot send a message to an encrypted room: %v", err)
	}
}
