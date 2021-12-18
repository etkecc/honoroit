package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/logger"
)

const (
	// ThreadRelation uses hardcoded value of element clients, should be replaced to m.thread after the MSC3440 release,
	// ref: https://github.com/matrix-org/matrix-doc/pull/3440/files#diff-113727ce0257b4dc0ad6f1087b6402f2cfcb6ff93272757b947bf1ce444056aeR296
	ThreadRelation = "io.element.thread"

	// TypingTimeout in milliseconds, used to avoid stuck typing status
	TypingTimeout = 5_000

	accountDataPrefix    = "cc.etke.honoroit."
	accountDataRooms     = accountDataPrefix + "rooms"
	accountDataSyncToken = accountDataPrefix + "batch_token"
)

// Bot represents matrix bot
type Bot struct {
	txt    *Text
	log    *logger.Logger
	api    *mautrix.Client
	cache  Cache
	name   string
	userID id.UserID
	roomID id.RoomID
}

// Config represents matrix config
type Config struct {
	// Homeserver url
	Homeserver string
	// Login is a localpart (honoroit - OK, @honoroit:example.com - wrong)
	Login string
	// Password for login/password auth only
	Password string
	// Token for access token auth only (not implemented yet)
	Token string
	// RoomID where threads will be created
	RoomID string
	// LogLevel for logger
	LogLevel string

	// Text messages
	Text *Text

	// Cache client
	Cache Cache
}

// Text messages
type Text struct {
	// Greetings message sent to customer on first contact
	Greetings string
	// Error message sent to customer if something goes wrong
	Error string
	// EmptyRoom message sent to backoffice/threads room when customer left his room
	EmptyRoom string
	// Done message sent to customer when request marked as done in the threads room
	Done string
}

// Cache client interface
type Cache interface {
	Set(string, interface{})
	Get(string) interface{}
}

// NewBot creates a new matrix bot
func NewBot(cfg *Config) (*Bot, error) {
	logger := logger.New("matrix.", cfg.LogLevel)
	apiBot, err := mautrix.NewClient(cfg.Homeserver, "", "")
	if err != nil {
		return nil, err
	}

	client := &Bot{
		api:   apiBot,
		log:   logger,
		txt:   cfg.Text,
		cache: cfg.Cache,
	}

	err = client.login(cfg.Login, cfg.Password)
	if err != nil {
		return nil, err
	}
	client.userID = client.api.UserID
	client.roomID = id.RoomID(cfg.RoomID)

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

// Start performs matrix /sync
func (b *Bot) Start() error {
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.StateMember, b.onInvite)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventMessage, b.onMessage)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventEncrypted, b.onEncryptedMessage)

	nameResp, err := b.api.GetOwnDisplayName()
	if err != nil {
		return err
	}
	b.name = nameResp.DisplayName

	err = b.api.SetPresence(event.PresenceOnline)
	if err != nil {
		return err
	}

	b.log.Info("bot has been started")
	return b.api.Sync()
}

// Stop the bot
func (b *Bot) Stop() {
	b.log.Debug("stopping the bot")
	err := b.api.SetPresence(event.PresenceOffline)
	if err != nil {
		b.log.Error("cannot set presence to offile: %v", err)
	}

	_, err = b.api.Logout()
	if err != nil {
		b.log.Error("cannot logout: %v", err)
	}

	b.log.Info("bot has been stopped")
}
