package matrix

import (
	"database/sql"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
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
)

// Bot represents matrix bot
type Bot struct {
	txt    *Text
	log    *logger.Logger
	api    *mautrix.Client
	olm    *crypto.OlmMachine
	store  *store.Store
	prefix string
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
	// RoomID where threads will be created
	RoomID string
	// Prefix of commands
	Prefix string
	// LogLevel for logger
	LogLevel string

	// Text messages
	Text *Text

	// DB object
	DB *sql.DB
	// Dialect of the DB: postgres, sqlite3
	Dialect string
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

// NewBot creates a new matrix bot
func NewBot(cfg *Config) (*Bot, error) {
	api, err := mautrix.NewClient(cfg.Homeserver, "", "")
	if err != nil {
		return nil, err
	}
	api.Logger = logger.New("api.", cfg.LogLevel)

	client := &Bot{
		api:    api,
		log:    logger.New("matrix.", cfg.LogLevel),
		txt:    cfg.Text,
		prefix: cfg.Prefix,
		roomID: id.RoomID(cfg.RoomID),
	}

	storer := store.New(cfg.DB, cfg.Dialect, logger.New("store.", cfg.LogLevel))
	if err = storer.CreateTables(); err != nil {
		return nil, err
	}
	client.store = storer
	client.api.Store = storer

	if err = client.login(cfg.Login, cfg.Password); err != nil {
		return nil, err
	}
	return client, nil
}

// WithEncryption adds OLM machine to the bot
func (b *Bot) WithEncryption() error {
	storeLog := logger.New("store.", b.log.GetLevel())
	cryptoLog := logger.New("olm.", b.log.GetLevel())
	if err := b.store.WithCrypto(b.api.UserID, b.api.DeviceID, storeLog); err != nil {
		return err
	}
	b.olm = crypto.NewOlmMachine(b.api, cryptoLog, b.store, b.store)

	return b.olm.Load()
}

// Start performs matrix /sync
func (b *Bot) Start() error {
	b.initSync()
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
	b.api.StopSync()
	b.log.Info("bot has been stopped")
}
