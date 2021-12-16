package matrix

import (
	"sync"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

const (
	// ThreadRelation uses hardcoded value of element clients, should be replaced to m.thread after the MSC3440 release,
	// ref: https://github.com/matrix-org/matrix-doc/pull/3440/files#diff-113727ce0257b4dc0ad6f1087b6402f2cfcb6ff93272757b947bf1ce444056aeR296
	ThreadRelation       = "io.element.thread"
	accountDataPrefix    = "cc.etke.honoroit."
	accountDataRooms     = accountDataPrefix + "rooms"
	accountDataSyncToken = accountDataPrefix + "batch_token"
)

// Bot represents matrix bot
type Bot struct {
	admu     sync.Mutex
	api      *mautrix.Client
	userID   id.UserID
	roomID   id.RoomID
	roomsMap *accountDataRoomsMap
}

// NewBot creates a new matrix bot
func NewBot(homeserver, username, password, roomID string) (*Bot, error) {
	apiBot, err := mautrix.NewClient(homeserver, "", "")
	if err != nil {
		return nil, err
	}

	client := &Bot{
		api: apiBot,
	}

	err = client.login(username, password)
	if err != nil {
		return nil, err
	}
	client.userID = client.api.UserID
	client.roomID = id.RoomID(roomID)

	return client, nil
}

// WithStore adds persistent storage to the bot. Right now it uses account data store, but will be changed in future
func (b *Bot) WithStore() error {
	filter := b.api.Syncer.GetFilterJSON(b.userID)
	filter.AccountData = mautrix.FilterPart{
		Limit: 50,
		NotTypes: []event.Type{
			event.NewEventType(accountDataSyncToken),
		},
	}
	filterResp, err := b.api.CreateFilter(filter)
	if err != nil {
		return err
	}

	b.api.Store = mautrix.NewAccountDataStore(accountDataSyncToken, b.api)
	b.api.Store.SaveFilterID(b.userID, filterResp.FilterID)

	return nil
}

// Run performs matrix /sync
func (b *Bot) Run() error {
	// preload mappings
	if err := b.loadRoomsMap(); err != nil {
		return err
	}

	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.StateMember, b.onInvite)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventMessage, b.onMessage)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventEncrypted, b.onEncryptedMessage)

	go b.syncRoomsMap()
	return b.api.Sync()
}
