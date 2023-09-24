package matrix

import (
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type RespThreads struct {
	Chunk     []*event.Event `json:"chunk"`
	NextBatch string         `json:"next_batch"`
}

// Threads returns all threads of the specified room
func (b *Bot) Threads(roomID id.RoomID, from string) (*RespThreads, error) {
	query := map[string]string{
		"from":  from,
		"limit": "100",
	}

	var resp *RespThreads
	urlPath := b.lp.GetClient().BuildURLWithQuery(mautrix.ClientURLPath{"v1", "rooms", roomID, "threads"}, query)
	_, err := b.lp.GetClient().MakeRequest("GET", urlPath, nil, &resp)
	return resp, err
}

func (b *Bot) countCustomerRequests(userID id.UserID) (int, int, error) {
	var user int
	var hs int
	var from string
	for {
		resp, err := b.Threads(b.roomID, from)
		if err != nil {
			b.log.Error().Err(err).Str("from", from).Str("roomID", b.roomID.String()).Msg("cannot request threads for the room")
			return user, hs, err
		}
		for _, evt := range resp.Chunk {
			if b.eventContains(evt, "customer", userID.String()) {
				user++
			}
			if b.eventContains(evt, "homeserver", userID.Homeserver()) {
				hs++
			}
		}
		from = resp.NextBatch
		if resp.NextBatch == "" {
			break
		}
	}

	return user, hs, nil
}

func (b *Bot) getName(userID id.UserID) string {
	name := userID.String()
	dnresp, err := b.lp.GetClient().GetDisplayName(userID)
	if err != nil {
		b.log.Warn().Err(err).Str("userID", userID.String()).Msg("cannot get display name")
	}
	if dnresp != nil {
		name = dnresp.DisplayName + " (" + name + ")"
	}

	return name
}

func (b *Bot) eventContains(evt *event.Event, attrName, attrValue string) bool {
	if evt.Content.Raw == nil {
		return false
	}
	if evt.Content.Raw[attrName] != attrValue {
		return false
	}

	return true
}
