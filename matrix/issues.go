package matrix

import (
	"context"
	"fmt"
	"strconv"

	redminelib "github.com/nixys/nxs-go-redmine/v5"
	"gitlab.com/etke.cc/honoroit/redmine"
	"gitlab.com/etke.cc/linkpearl"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
)

const issueNotePrefix = "cc.etke.honoroit.redmine.note."

// SyncIssues fetches redmine issue status and notes and sends them to the corresponding rooms
// due to the lack of a proper redmine webhook, this function is called periodically,
// due to the complexity of that function, it has own "syncing" flag that prevents multiple calls at once
func (b *Bot) SyncIssues() {
	if !b.redmine.Enabled() {
		return
	}

	ctx := context.Background()
	if b.syncing {
		b.log.Debug().Msg("already syncing")
		return
	}
	b.log.Debug().Msg("syncing")
	b.syncing = true
	defer func() { b.syncing = false }()

	threadIDs := b.getThreadIDs(ctx)
	for _, threadID := range threadIDs {
		threadID := threadID
		roomID, err := b.findRoomID(ctx, threadID)
		if err != nil || roomID == "" {
			continue
		}

		issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
		if err != nil {
			continue
		}
		issueID, err := strconv.Atoi(issueIDStr)
		if err != nil || issueID == 0 {
			continue
		}
		b.syncIssueNotes(ctx, threadID, roomID, issueID)
		b.syncIssueStatus(ctx, threadID, issueID)
	}
}

func (b *Bot) syncIssueStatus(ctx context.Context, threadID id.EventID, issueID int) {
	closed, err := b.redmine.IsClosed(int64(issueID))
	if err != nil {
		b.log.Error().Err(err).Msg("cannot get redmine status")
		return
	}
	if !closed {
		return
	}
	content := format.RenderMarkdown("_closed from redmine_", true, true)
	content.RelatesTo = linkpearl.RelatesTo(threadID)
	fullContent := event.Content{
		Parsed: &content,
	}

	b.closeRequest(ctx, &event.Event{Content: fullContent}, false)
}

func (b *Bot) syncIssueNotes(ctx context.Context, threadID id.EventID, roomID id.RoomID, issueID int) {
	notes, err := b.redmine.GetNotes(int64(issueID))
	if err != nil {
		b.log.Error().Err(err).Msg("cannot get redmine notes")
		return
	}
	for _, note := range notes {
		acID := issueNotePrefix + threadID.String() + "_" + strconv.Itoa(int(note.ID))
		data, err := b.lp.GetAccountData(ctx, acID)
		if err != nil {
			b.log.Error().Err(err).Msg("cannot get account data")
			return
		}
		if data["synced"] != "" {
			b.log.Debug().Str("acID", acID).Msg("note already synced")
			return
		}
		b.syncIssueNote(ctx, threadID, roomID, acID, note)
	}
}

func (b *Bot) syncIssueNote(ctx context.Context, threadID id.EventID, roomID id.RoomID, acID string, note *redminelib.IssueJournalObject) {
	// first, send the note to the thread
	var prefix string
	if note.PrivateNotes {
		prefix = fmt.Sprintf("_synced redmine private note #%d_", note.ID)
	} else {
		prefix = fmt.Sprintf("_synced redmine note #%d_", note.ID)
	}
	content := format.RenderMarkdown(fmt.Sprintf("%s\n\n%s", prefix, note.Notes), true, true)
	content.MsgType = event.MsgNotice
	content.RelatesTo = linkpearl.RelatesTo(threadID)
	evtID, err := b.lp.Send(ctx, b.roomID, content)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send message")
		return
	}

	if !note.PrivateNotes {
		// then, send the note to the customer's room
		content = format.RenderMarkdown(note.Notes, true, true)
		b.clearReply(&content)
		fullContent := &event.Content{
			Parsed: content,
			Raw: map[string]any{
				"event_id": evtID,
			},
		}
		_, err = b.lp.Send(ctx, roomID, fullContent)
		if err != nil {
			b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, linkpearl.RelatesTo(threadID))
			return
		}
	}
	if err := b.lp.SetAccountData(ctx, acID, map[string]string{"synced": "true"}); err != nil {
		b.log.Error().Err(err).Msg("cannot set account data")
	}
}

func (b *Bot) updateIssue(ctx context.Context, threadID id.EventID, text string) {
	key := "redmine_" + threadID.String()
	b.lock(key)
	defer b.unlock(key)

	issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
	if err != nil {
		return
	}
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil || issueID == 0 {
		return
	}
	if updateErr := b.redmine.UpdateIssue(int64(issueID), redmine.InProgress, text); updateErr != nil {
		b.log.Error().Err(updateErr).Msg("cannot update redmine issue")
	}
}

func (b *Bot) closeIssue(ctx context.Context, roomID id.RoomID, threadID id.EventID, text string) {
	key := "redmine_" + threadID.String()
	b.lock(key)
	defer b.unlock(key)

	issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
	if err != nil {
		return
	}
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil || issueID == 0 {
		return
	}
	if updateErr := b.redmine.UpdateIssue(int64(issueID), redmine.Done, text); updateErr != nil {
		b.log.Error().Err(updateErr).Msg("cannot close redmine issue")
		return
	}
	b.removeRedmineMapping(ctx, threadID.String())
	b.removeRedmineMapping(ctx, roomID.String())
	b.removeRedmineMapping(ctx, issueIDStr)
}
