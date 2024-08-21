package matrix

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/etkecc/go-linkpearl"
	"github.com/etkecc/go-mxidwc"
	"github.com/etkecc/go-psd"
	"github.com/etkecc/go-redmine"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/id"

	"github.com/etkecc/honoroit/internal/matrix/config"
)

// TypingTimeout in milliseconds, used to avoid stuck typing status
const TypingTimeout = 5_000

// Bot represents matrix bot
type Bot struct {
	cfg           *config.Manager
	log           *zerolog.Logger
	psdc          *psd.Client
	redmine       *redmine.Redmine
	lp            *linkpearl.Linkpearl
	mu            map[string]*sync.Mutex
	syncing       bool
	namesCache    *lru.Cache[id.UserID, [2]string]
	profilesCache *lru.Cache[id.UserID, *MSC4144Profile]
	eventsCache   *lru.Cache[id.EventID, id.EventID]
	prefix        string
	roomID        id.RoomID
	ignoreBefore  int64 // TODO remove after some time
}

// NewBot creates a new matrix bot
func NewBot(
	lp *linkpearl.Linkpearl,
	log *zerolog.Logger,
	cfg *config.Manager,
	psdc *psd.Client,
	rdm *redmine.Redmine,
	prefix string,
	roomID string,
	cacheSize int,
) (*Bot, error) {
	namesCache, err := lru.New[id.UserID, [2]string](cacheSize)
	if err != nil {
		return nil, err
	}
	profilesCache, err := lru.New[id.UserID, *MSC4144Profile](cacheSize)
	if err != nil {
		return nil, err
	}
	eventsCache, err := lru.New[id.EventID, id.EventID](cacheSize)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		lp:            lp,
		mu:            make(map[string]*sync.Mutex),
		cfg:           cfg,
		log:           log,
		psdc:          psdc,
		redmine:       rdm,
		namesCache:    namesCache,
		profilesCache: profilesCache,
		eventsCache:   eventsCache,
		prefix:        prefix,
		roomID:        id.RoomID(roomID),
	}
	bot.ignoreBefore = bot.cfg.Mautrix015Migration(context.Background())

	return bot, nil
}

// TODO: make it configurable
func (b *Bot) AutoCloseRequests() {
	ctx := context.Background()
	maxTS := time.Now().UTC().Add(-7 * 24 * time.Hour).UnixMilli()
	threadIDs := b.getThreadIDs(ctx)
	if len(threadIDs) == 0 {
		b.log.Info().Msg("no threads to close")
		return
	}

	for _, threadID := range threadIDs {
		// not an actual support thread - ignore
		if _, err := b.findRoomID(ctx, threadID); err != nil {
			continue
		}
		lastEvt := b.getLastThreadMessage(ctx, threadID)
		if lastEvt == nil {
			b.log.Info().Any("threadID", threadID).Msg("no last event fount")
			continue
		}
		// set relates_to to the thread
		lastEvt.RoomID = b.roomID
		content := lastEvt.Content.AsMessage()
		content.RelatesTo = linkpearl.RelatesTo(threadID)
		lastEvt.Content.Parsed = content

		// if last event timestamp is older than 7d - close the thread
		if lastEvt.Timestamp < maxTS {
			b.closeRequest(ctx, lastEvt, true)
		}
	}
}

// Start performs matrix /sync
func (b *Bot) Start() error {
	b.initSync()
	return b.lp.Start(context.Background())
}

// Stop the bot
func (b *Bot) Stop() {
	b.lp.Stop(context.Background())
}

func parseMXIDpatterns(patternsString string) ([]*regexp.Regexp, error) {
	defaultPattern := "@*:*"
	patterns := strings.Split(patternsString, ",")

	if len(patterns) == 0 && defaultPattern != "" {
		patterns = []string{defaultPattern}
	}

	return mxidwc.ParsePatterns(patterns)
}
