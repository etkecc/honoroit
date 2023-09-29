package matrix

import (
	"regexp"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog"
	"gitlab.com/etke.cc/go/mxidwc"
	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/matrix/config"
)

// TypingTimeout in milliseconds, used to avoid stuck typing status
const TypingTimeout = 5_000

// Bot represents matrix bot
type Bot struct {
	cfg          *config.Manager
	log          *zerolog.Logger
	lp           *linkpearl.Linkpearl
	mu           map[string]*sync.Mutex
	eventsCache  *lru.Cache
	prefix       string
	roomID       id.RoomID
	ignoreBefore int64 // TODO remove after some time
}

// NewBot creates a new matrix bot
func NewBot(
	lp *linkpearl.Linkpearl,
	log *zerolog.Logger,
	cfg *config.Manager,
	prefix string,
	roomID string,
	cacheSize int,
) (*Bot, error) {
	cache, err := lru.New(cacheSize)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		lp:          lp,
		mu:          make(map[string]*sync.Mutex),
		cfg:         cfg,
		log:         log,
		eventsCache: cache,
		prefix:      prefix,
		roomID:      id.RoomID(roomID),
	}
	bot.ignoreBefore = bot.cfg.Mautrix015Migration()

	return bot, nil
}

// Start performs matrix /sync
func (b *Bot) Start() error {
	b.migrateMappings()
	b.initSync()
	return b.lp.Start()
}

// Stop the bot
func (b *Bot) Stop() {
	b.lp.Stop()
}

func parseMXIDpatterns(patternsString string) ([]*regexp.Regexp, error) {
	defaultPattern := "@*:*"
	patterns := strings.Split(patternsString, ",")

	if len(patterns) == 0 && defaultPattern != "" {
		patterns = []string{defaultPattern}
	}

	return mxidwc.ParsePatterns(patterns)
}
