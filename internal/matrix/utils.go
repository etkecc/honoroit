package matrix

import (
	"context"

	"github.com/etkecc/go-linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

// SendNotice is a copy of linkpearl.SendNotice, but with raw content support
func (b *Bot) SendNotice(ctx context.Context, roomID id.RoomID, message string, raw map[string]any, relates ...*event.RelatesTo) id.EventID {
	var relatesTo *event.RelatesTo
	if len(relates) > 0 {
		relatesTo = relates[0]
	}
	content := format.RenderMarkdown(message, true, true)
	content.MsgType = event.MsgNotice
	content.RelatesTo = relatesTo
	fullContent := &event.Content{Parsed: &content, Raw: raw}

	eventID, err := b.lp.Send(ctx, roomID, fullContent)
	if err != nil {
		b.log.Error().Err(linkpearl.UnwrapError(err)).Str("roomID", roomID.String()).Str("retries", "1/2").Msg("cannot send a notice into the room")
		if relatesTo != nil {
			content.RelatesTo = nil
			fullContent.Parsed = &content
			eventID, err = b.lp.Send(ctx, roomID, fullContent)
			if err != nil {
				b.log.Error().Err(linkpearl.UnwrapError(err)).Str("roomID", roomID.String()).Str("retries", "2/2").Msg("cannot send a notice into the room even without relations")
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
