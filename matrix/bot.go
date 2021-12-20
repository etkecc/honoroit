package matrix

import (
	"database/sql"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/logger"
	"gitlab.com/etke.cc/honoroit/matrix/store"
)

const (
	// ThreadRelation uses hardcoded value of element clients, should be replaced to m.thread after the MSC3440 release,
	// ref: https://github.com/matrix-org/matrix-doc/pull/3440/files#diff-113727ce0257b4dc0ad6f1087b6402f2cfcb6ff93272757b947bf1ce444056aeR296
	ThreadRelation = "io.element.thread"

	// TypingTimeout in milliseconds, used to avoid stuck typing status
	TypingTimeout = 5_000

	accountDataPrefix       = "cc.etke.honoroit."
	accountDataRooms        = accountDataPrefix + "rooms"
	accountDataSessionToken = accountDataPrefix + "session_token"
)

// Bot represents matrix bot
type Bot struct {
	txt      *Text
	log      *logger.Logger
	api      *mautrix.Client
	store    *store.Store
	cache    Cache
	name     string
	userID   id.UserID
	deviceID id.DeviceID
	roomID   id.RoomID
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
	api, err := mautrix.NewClient(cfg.Homeserver, "", cfg.Token)
	if err != nil {
		return nil, err
	}

	client := &Bot{
		api:    api,
		log:    logger,
		txt:    cfg.Text,
		cache:  cfg.Cache,
		roomID: id.RoomID(cfg.RoomID),
	}

	if cfg.Token == "" {
		if err = client.login(cfg.Login, cfg.Password); err != nil {
			return nil, err
		}
	}

	if err = client.hydrate(); err != nil {
		return nil, err
	}

	return client, nil
}

// WithStore adds persistent storage to the bot.
func (b *Bot) WithStore(db *sql.DB, dialect string) error {
	// MIGRATION. TODO: remove
	key := "cc.etke.honoroit.batch_token"
	type accountData struct {
		NextBatch string
	}
	data := accountData{}
	// nolint // if there is a error, that means migration already done
	b.api.GetAccountData(key, &data)
	// nolint // if there is a error, that means migration already done
	b.api.SetAccountData(key, accountData{})

	cfg := &store.Config{
		DB:      db,
		Dialect: dialect,
		Logger:  logger.New("store.", b.log.GetLevel()),
	}
	storer := store.New(cfg)
	err := storer.CreateTables()
	if err != nil {
		return err
	}

	b.store = storer
	b.api.Store = storer
	return nil
}

// Start performs matrix /sync
func (b *Bot) Start() error {
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.StateMember, b.onMembership)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventMessage, b.onMessage)
	b.api.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventEncrypted, b.onEncryptedMessage)

	err := b.api.SetPresence(event.PresenceOnline)
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

	b.log.Info("bot has been stopped")
}
