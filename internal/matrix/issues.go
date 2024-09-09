package matrix

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/etkecc/go-linkpearl"
	"github.com/etkecc/go-redmine"
	redminelib "github.com/nixys/nxs-go-redmine/v5"
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
		b.log.Debug().Msg("redmine is disabled, skipping sync")
		return
	}

	ctx := context.Background()
	if b.syncing {
		b.log.Debug().Msg("already syncing redmine issues")
		return
	}
	b.log.Debug().Msg("syncing redmine issues")
	b.syncing = true
	defer func() { b.syncing = false }()

	threadIDs := b.getThreadIDs(ctx)
	for _, threadID := range threadIDs {
		b.syncIssue(ctx, threadID)
	}
}

func (b *Bot) syncIssue(ctx context.Context, threadID id.EventID) {
	roomID, err := b.findRoomID(ctx, threadID)
	if err != nil || roomID == "" {
		if err != nil && !errors.Is(err, errNotMapped) {
			b.log.Warn().Err(err).Str("thread_id", threadID.String()).Msg("cannot find room")
		}
		return
	}

	issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
	if err != nil {
		if !errors.Is(err, errNotMapped) {
			b.log.Warn().Err(err).Str("thread_id", threadID.String()).Msg("cannot get redmine mapping")
		}
		b.log.Debug().Str("thread_id", threadID.String()).Str("room_id", roomID.String()).Msg("request not mapped")
		return
	}
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil || issueID == 0 {
		b.log.Debug().Str("thread_id", threadID.String()).Str("room_id", roomID.String()).Msg("issue not found")
		return
	}
	b.log.Debug().Str("thread_id", threadID.String()).Str("room_id", roomID.String()).Int("issue_id", issueID).Msg("syncing request")
	b.syncIssueNotes(ctx, threadID, roomID, issueID)
	b.syncIssueStatus(ctx, threadID, issueID)
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
	b.log.Debug().Int("issue_id", issueID).Int("notes", len(notes)).Msg("syncing notes")
	for _, note := range notes {
		acID := issueNotePrefix + threadID.String() + "_" + strconv.Itoa(int(note.ID))
		data, err := b.lp.GetAccountData(ctx, acID)
		if err != nil {
			b.log.Error().Err(err).Msg("cannot get account data")
			continue
		}
		if data["synced"] != "" {
			b.log.Debug().Int("issue_id", issueID).Int("note_id", int(note.ID)).Msg("note already synced")
			continue
		}
		b.log.Debug().Int("issue_id", issueID).Int("note_id", int(note.ID)).Msg("syncing note")
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
	b.log.Debug().Str("thread_id", threadID.String()).Int("note_id", int(note.ID)).Msg("sending note to operators room")
	evtID, err := b.lp.Send(ctx, b.roomID, content)
	if err != nil {
		b.log.Error().Err(err).Msg("cannot send note to operators room")
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
		b.log.Debug().Str("thread_id", threadID.String()).Int("note_id", int(note.ID)).Msg("sending note to customer room")
		_, err = b.lp.Send(ctx, roomID, fullContent)
		if err != nil {
			b.log.Error().Err(err).Msg("cannot send note to customer room")
			b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, linkpearl.RelatesTo(threadID))
			return
		}
	}
	b.log.Debug().Str("thread_id", threadID.String()).Int("note_id", int(note.ID)).Msg("marking note as synced")
	if err := b.lp.SetAccountData(ctx, acID, map[string]string{"synced": "true"}); err != nil {
		b.log.Error().Err(err).Msg("cannot set account data")
	}
}

// getFileUploadReq returns a redmine.UploadRequest for the given content (if it's a file)
func (b *Bot) getFileUploadReq(ctx context.Context, content *event.MessageEventContent) *redmine.UploadRequest {
	var fileEncrypted bool
	fileName, fileMXCURL := GetFileURL(content)
	if fileMXCURL == "" {
		return nil
	}

	if content.GetFile().URL == fileMXCURL {
		fileEncrypted = true
	}
	fileContentURI, err := fileMXCURL.Parse()
	if err != nil {
		b.log.Warn().Err(err).Str("filename", fileName).Str("url", string(fileMXCURL)).Msg("cannot parse file MXC URI")
		return nil
	}

	if fileContentURI.IsEmpty() {
		return nil
	}

	resp, err := b.lp.GetClient().Download(ctx, fileContentURI)
	if err != nil {
		b.log.Warn().Err(err).Str("filename", fileName).Str("url", string(fileMXCURL)).Msg("cannot download file")
		return nil
	}

	if !fileEncrypted {
		return &redmine.UploadRequest{
			Path:   fileName,
			Stream: resp.Body,
		}
	}

	defer resp.Body.Close()
	if err := content.File.PrepareForDecryption(); err != nil {
		b.log.Warn().Err(err).Msg("cannot prepare file for decryption")
		return nil
	}
	encStream := content.File.DecryptStream(resp.Body)
	return &redmine.UploadRequest{
		Path:   fileName,
		Stream: encStream,
	}
}

func (b *Bot) updateIssue(ctx context.Context, byOperator bool, sender string, threadID id.EventID, content *event.MessageEventContent) {
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
	var statusID int64
	var text string
	if byOperator {
		statusID = b.redmine.StatusToID(redmine.WaitingForCustomer)
		text = fmt.Sprintf("_%s (👩‍💼 operator)_\n\n%s", sender, content.Body)
	} else {
		statusID = b.redmine.StatusToID(redmine.WaitingForOperator)
		text = fmt.Sprintf("_%s (🧑‍🦱customer)_\n\n%s", sender, content.Body)
	}

	if updateErr := b.redmine.UpdateIssue(int64(issueID), statusID, text, b.getFileUploadReq(ctx, content)); updateErr != nil {
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
	log := b.log.With().Int("issue_id", issueID).Logger()
	issue, err := b.redmine.GetIssue(int64(issueID), "attachments")
	if err != nil {
		log.Warn().Err(err).Msg("cannot get redmine issue")
		return
	}
	if issue.Attachments != nil {
		//nolint:gocritic // we can't do anything with it
		for _, attachment := range *issue.Attachments {
			if err := b.redmine.DeleteAttachment(attachment.ID); err != nil {
				log.Warn().Err(err).Int("attachment_id", int(attachment.ID)).Msg("cannot delete attachment")
			}
		}
	}
	if updateErr := b.redmine.UpdateIssue(int64(issueID), b.redmine.StatusToID(redmine.Done), text); updateErr != nil {
		log.Warn().Err(updateErr).Msg("cannot close redmine issue")
		return
	}
	b.removeRedmineMapping(ctx, threadID.String())
	b.removeRedmineMapping(ctx, roomID.String())
	b.removeRedmineMapping(ctx, issueIDStr)
}
