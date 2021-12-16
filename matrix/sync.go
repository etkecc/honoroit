package matrix

import (
	"fmt"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

func (b *Bot) onInvite(_ mautrix.EventSource, evt *event.Event) {
	userID := b.userID.String()
	invite := evt.Content.AsMember().Membership == event.MembershipInvite

	if invite && evt.GetStateKey() == userID {
		_, err := b.api.JoinRoomByID(evt.RoomID)
		if err != nil {
			fmt.Println("cannot join room:", err)
		}
	}
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
		b.error(b.roomID, "cannot send a message to an encrypted room: "+err.Error())
	}
}
