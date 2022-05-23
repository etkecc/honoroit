package linkpearl

import (
	"time"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

// OnEventType allows callers to be notified when there are new events for the given event type.
// There are no duplicate checks.
func (l *Linkpearl) OnEventType(eventType event.Type, callback mautrix.EventHandler) {
	l.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(eventType, callback)
}

// OnSync shortcut to mautrix.DefaultSyncer.OnSync
func (l *Linkpearl) OnSync(callback mautrix.SyncHandler) {
	l.api.Syncer.(*mautrix.DefaultSyncer).OnSync(callback)
}

// OnEvent shortcut to mautrix.DefaultSyncer.OnEvent
func (l *Linkpearl) OnEvent(callback mautrix.EventHandler) {
	l.api.Syncer.(*mautrix.DefaultSyncer).OnEvent(callback)
}

func (l *Linkpearl) initSync() {
	l.api.Syncer.(*mautrix.DefaultSyncer).OnSync(l.olm.ProcessSyncResponse)
	l.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.StateEncryption,
		func(source mautrix.EventSource, evt *event.Event) {
			go l.onEncryption(source, evt)
		},
	)

	l.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(
		event.StateMember,
		func(source mautrix.EventSource, evt *event.Event) {
			go l.onMembership(source, evt)
		},
	)
}

func (l *Linkpearl) onMembership(_ mautrix.EventSource, evt *event.Event) {
	l.olm.HandleMemberEvent(evt)
	l.store.SetMembership(evt)

	// autoaccept invites
	l.onInvite(evt, 0)
	// autoleave empty rooms
	l.onEmpty(evt)
}

func (l *Linkpearl) onInvite(evt *event.Event, retry int) {
	userID := l.api.UserID.String()
	invite := evt.Content.AsMember().Membership == event.MembershipInvite
	if invite && evt.GetStateKey() == userID {
		_, err := l.api.JoinRoomByID(evt.RoomID)
		if err != nil {
			l.log.Error("cannot join the room %s: %v", evt.RoomID, err)
			if retry < l.maxretries {
				time.Sleep(5 * time.Second)
				l.log.Debug("trying to join again (%d/%d)", retry+1, l.maxretries)
				l.onInvite(evt, retry+1)
			}
		}
	}
}

func (l *Linkpearl) onEmpty(evt *event.Event) {
	if !l.autoleave {
		return
	}

	members := l.store.GetRoomMembers(evt.RoomID)
	if len(members) >= 1 && members[0] != l.api.UserID {
		return
	}

	_, err := l.api.LeaveRoom(evt.RoomID)
	if err != nil {
		l.log.Error("cannot leave room: %v", err)
	}
	l.log.Debug("left room %s because it's empty", evt.RoomID)
}

func (l *Linkpearl) onEncryption(_ mautrix.EventSource, evt *event.Event) {
	l.store.SetEncryptionEvent(evt)
}
