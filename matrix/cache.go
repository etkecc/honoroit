package matrix

import "maunium.net/go/mautrix/id"

const (
	cacheEventPrefix  = "event_"
	cacheThreadPrefix = "thread_"
)

func (b *Bot) getCache(eventID id.EventID) id.EventID {
	keys := []string{cacheEventPrefix + string(eventID), cacheThreadPrefix + string(eventID)}
	var threadID id.EventID

	for _, key := range keys {
		if !b.cache.Has(key) {
			continue
		}

		threadID = b.cache.Get(key).(id.EventID)
		b.log.Debug("threadID %s found in cache", threadID)
		break
	}
	return threadID
}

func (b *Bot) setCache(eventID id.EventID, value id.EventID) {
	keys := []string{cacheEventPrefix + string(eventID), cacheThreadPrefix + string(value)}
	for _, key := range keys {
		if b.cache.Has(key) {
			continue
		}

		b.log.Debug("saving cache for %s = %s", key, value)
		b.cache.Set(key, value)
	}
}
