package matrix

import "maunium.net/go/mautrix/id"

const (
	cacheEventPrefix  = "event_"
	cacheThreadPrefix = "thread_"
)

func (b *Bot) getCache(eventID id.EventID) id.EventID {
	var ok bool
	var threadID id.EventID
	keys := []string{cacheEventPrefix + string(eventID), cacheThreadPrefix + string(eventID)}

	for _, key := range keys {
		if !b.cache.Has(key) {
			continue
		}

		value := b.cache.Get(key)
		if value == nil {
			continue
		}

		threadID, ok = value.(id.EventID)
		if threadID == "" || !ok {
			continue
		}

		b.log.Debug("threadID %s found in cache", threadID)
		break
	}
	return threadID
}

func (b *Bot) setCache(eventID id.EventID, value id.EventID) {
	keys := []string{cacheEventPrefix + string(eventID), cacheThreadPrefix + string(value)}
	for _, key := range keys {
		b.log.Debug("saving cache for %s = %s", key, value)
		b.cache.Set(key, value)
	}
}
