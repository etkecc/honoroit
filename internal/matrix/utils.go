package matrix

import (
	"context"
	"strings"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"

	"github.com/etkecc/honoroit/internal/matrix/config"
)

func (b *Bot) parseContent(evt *event.Event) {
	if evt.Content.Parsed != nil {
		return
	}
	if err := evt.Content.ParseRaw(evt.Type); err != nil {
		b.log.Error("cannot parse event content", "error", err)
	}
}

// SendNotice is linkpearl.SendNotice plus raw content support and config-based msg type setting
func (b *Bot) SendNotice(ctx context.Context, roomID id.RoomID, message string, raw map[string]any, relates ...*event.RelatesTo) id.EventID {
	var relatesTo *event.RelatesTo
	if len(relates) > 0 {
		relatesTo = relates[0]
	}
	content := format.RenderMarkdown(message, true, true)
	content.MsgType = event.MessageType(b.cfg.Get(ctx, config.MsgType.Key))
	content.RelatesTo = relatesTo
	fullContent := &event.Content{Parsed: &content, Raw: raw}

	eventID, err := b.lp.Send(ctx, roomID, fullContent)
	if err != nil {
		b.log.Error("cannot send a notice into the room", "error", linkpearl.UnwrapError(err), "roomID", roomID.String(), "retries", "1/2")
		if relatesTo != nil {
			content.RelatesTo = nil
			fullContent.Parsed = &content
			eventID, err = b.lp.Send(ctx, roomID, fullContent)
			if err != nil {
				b.log.Error("cannot send a notice into the room even without relations", "error", linkpearl.UnwrapError(err), "roomID", roomID.String(), "retries", "2/2")
				return ""
			}
		}
	}
	return eventID
}

func GetFileURL(content *event.MessageEventContent) (fileName string, fileURL id.ContentURIString) {
	if content == nil {
		return "", ""
	}

	fileName = content.Body
	if content.FileName != "" {
		fileName = content.FileName
	}

	if content.URL != "" {
		return fileName, content.URL
	}

	if content.GetInfo().ThumbnailURL != "" {
		return fileName, content.GetInfo().ThumbnailURL
	}

	if content.GetFile().URL != "" {
		return fileName, content.GetFile().URL
	}
	return "", ""
}

// isRoomConfigured checks if the operator room is configured
func isRoomConfigured(roomID string) bool {
	if roomID == "" {
		return false
	}
	return strings.HasPrefix(roomID, "!")
}
