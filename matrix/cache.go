package matrix

import "maunium.net/go/mautrix/id"

const (
	cacheEventPrefix  = "event_"
	cacheThreadPrefix = "thread_"
	cacheNamePrefix   = "name_"
)

func (b *Bot) getCachedThreadID(eventID id.EventID) id.EventID {
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

func (b *Bot) setCachedThreadID(eventID id.EventID, value id.EventID) {
	keys := []string{cacheEventPrefix + string(eventID), cacheThreadPrefix + string(value)}
	for _, key := range keys {
		b.log.Debug("saving cache for %s = %s", key, value)
		b.cache.Set(key, value)
	}
}

func (b *Bot) getCachedName(userID id.UserID) string {
	key := cacheNamePrefix + userID.String()
	if !b.cache.Has(key) {
		return ""
	}

	value := b.cache.Get(key)
	if value == nil {
		return ""
	}

	name, ok := value.(string)
	if name == "" || !ok {
		return ""
	}

	b.log.Debug("name %s found in cache", name)
	return name
}

func (b *Bot) setCachedName(userID id.UserID, value string) {
	key := cacheNamePrefix + userID.String()
	b.log.Debug("saving cache for %s = %s", key, value)
	b.cache.Set(key, value)
}
