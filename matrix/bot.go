package matrix

import (
	"database/sql"
	"regexp"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"gitlab.com/etke.cc/go/logger"
	"gitlab.com/etke.cc/go/mxidwc"
	"gitlab.com/etke.cc/linkpearl"
	"gitlab.com/etke.cc/linkpearl/config"
	"maunium.net/go/mautrix/id"
)

const (
	// ThreadRelation with stable prefix
	ThreadRelation = "m.thread"
	// ThreadRelationOld uses hardcoded value of element clients, should be replaced to m.thread after the MSC3440 release,
	// ref: https://github.com/matrix-org/matrix-doc/pull/3440/files#diff-113727ce0257b4dc0ad6f1087b6402f2cfcb6ff93272757b947bf1ce444056aeR296
	ThreadRelationOld = "io.element.thread"

	// TypingTimeout in milliseconds, used to avoid stuck typing status
	TypingTimeout = 5_000
)

// Bot represents matrix bot
type Bot struct {
	txt            *Text
	log            *logger.Logger
	lp             *linkpearl.Linkpearl
	mu             map[string]*sync.Mutex
	eventsCache    *lru.Cache
	prefix         string
	prefixes       []string
	roomID         id.RoomID
	allowedUsers   []*regexp.Regexp
	ignoredRooms   map[id.RoomID]struct{}
	ignoreNoThread bool
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
	// AllowedUsers is list of wildcard rules to allow requests only from specific users
	AllowedUsers []string
	// IgnoredRooms list of room IDs to ignore
	IgnoredRooms []string
	// IgnoreNoThread mode completely ignores any messages sent outside of thread
	IgnoreNoThread bool
	// Prefix of commands
	Prefix string
	// LogLevel for logger
	LogLevel string
	// NoEncryption disabled matrix e2e encryption
	NoEncryption bool

	// Text messages
	Text *Text

	// DB object
	DB *sql.DB
	// Dialect of the DB: postgres, sqlite3
	Dialect string

	// Cache size
	CacheSize int
}

// Text messages
type Text struct {
	// PrefixOpen is a prefix added to new thread topics
	PrefixOpen string
	// PrefixDone is a prefix added to threads marked as done/closed
	PrefixDone string

	// NoEncryption message sent to customer when encryption disabled and customer tries to use encrypted chat
	NoEncryption string
	// Greetings message sent to customer on first contact
	Greetings string
	// Join message sent to backoffice/threads room when customer joins a room
	Join string
	// Invite message sent to backoffice/threads room when customer invites somebody into a room
	Invite string
	// Leave message sent to backoffice/threads room when a customer leaves a room
	Leave string
	// Error message sent to customer if something goes wrong
	Error string
	// EmptyRoom message sent to backoffice/threads room when the last customer left a room
	EmptyRoom string
	// Start message that sent into the read as result of the "start" command
	Start string
	// Done message sent to customer when request marked as done in the threads room
	Done string
}

// NewBot creates a new matrix bot
func NewBot(cfg *Config) (*Bot, error) {
	log := logger.New("matrix.", cfg.LogLevel)
	lp, err := linkpearl.New(&config.Config{
		Homeserver:   cfg.Homeserver,
		Login:        cfg.Login,
		Password:     cfg.Password,
		DB:           cfg.DB,
		Dialect:      cfg.Dialect,
		LPLogger:     log,
		APILogger:    logger.New("api.", cfg.LogLevel),
		StoreLogger:  logger.New("store.", cfg.LogLevel),
		CryptoLogger: logger.New("olm.", cfg.LogLevel),
		NoEncryption: cfg.NoEncryption,
	})
	if err != nil {
		return nil, err
	}

	ignoredRoomIDs := make(map[id.RoomID]struct{}, len(cfg.IgnoredRooms))
	for _, room := range cfg.IgnoredRooms {
		ignoredRoomIDs[id.RoomID(room)] = struct{}{}
	}

	allowedUsers, uerr := parseMXIDpatterns(cfg.AllowedUsers, "@*:*")
	if uerr != nil {
		return nil, uerr
	}
	cache, err := lru.New(cfg.CacheSize)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		lp:             lp,
		mu:             make(map[string]*sync.Mutex),
		log:            log,
		txt:            cfg.Text,
		eventsCache:    cache,
		prefix:         cfg.Prefix,
		prefixes:       []string{cfg.Text.PrefixOpen, cfg.Text.PrefixDone},
		roomID:         id.RoomID(cfg.RoomID),
		allowedUsers:   allowedUsers,
		ignoredRooms:   ignoredRoomIDs,
		ignoreNoThread: cfg.IgnoreNoThread,
	}

	return bot, nil
}

// Start performs matrix /sync
func (b *Bot) Start() error {
	if err := b.migrate(); err != nil {
		return err
	}

	b.initSync()
	return b.lp.Start()
}

// Stop the bot
func (b *Bot) Stop() {
	b.lp.Stop()
}

func parseMXIDpatterns(patterns []string, defaultPattern string) ([]*regexp.Regexp, error) {
	if len(patterns) == 0 && defaultPattern != "" {
		patterns = []string{defaultPattern}
	}

	return mxidwc.ParsePatterns(patterns)
}
