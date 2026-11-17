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

// SyncIssues sends redmine status/notes to rooms periodically (no webhook), guarded against concurrent runs
func (b *Bot) SyncIssues() {
	if !b.redmine.Enabled() {
		b.log.Debug("redmine is disabled, skipping sync")
		return
	}

	ctx := context.Background()
	if b.syncing {
		b.log.Debug("already syncing redmine issues")
		return
	}
	b.log.Debug("syncing redmine issues")
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
			b.log.Warn("cannot find room", "error", err, "thread_id", threadID.String())
		}
		return
	}

	issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
	if err != nil {
		if !errors.Is(err, errNotMapped) {
			b.log.Warn("cannot get redmine mapping", "error", err, "thread_id", threadID.String())
		}
		b.log.Debug("request not mapped", "thread_id", threadID.String(), "room_id", roomID.String())
		return
	}
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil || issueID == 0 {
		b.log.Debug("issue not found", "thread_id", threadID.String(), "room_id", roomID.String())
		return
	}
	b.log.Debug("syncing request", "thread_id", threadID.String(), "room_id", roomID.String(), "issue_id", issueID)
	b.syncIssueNotes(ctx, threadID, roomID, issueID)
	b.syncIssueStatus(ctx, threadID, issueID)
}

func (b *Bot) syncIssueStatus(ctx context.Context, threadID id.EventID, issueID int) {
	closed, err := b.redmine.IsClosed(int64(issueID))
	if err != nil {
		b.log.Error("cannot get redmine status", "error", err)
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
		b.log.Error("cannot get redmine notes", "error", err)
		return
	}
	b.log.Debug("syncing notes", "issue_id", issueID, "notes", len(notes))
	for _, note := range notes {
		acID := issueNotePrefix + threadID.String() + "_" + strconv.Itoa(int(note.ID))
		data, err := b.lp.GetAccountData(ctx, acID)
		if err != nil {
			b.log.Error("cannot get account data", "error", err)
			continue
		}
		if data["synced"] != "" {
			b.log.Debug("note already synced", "issue_id", issueID, "note_id", int(note.ID))
			continue
		}
		b.log.Debug("syncing note", "issue_id", issueID, "note_id", int(note.ID))
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
	b.log.Debug("sending note to operators room", "thread_id", threadID.String(), "note_id", int(note.ID))
	evtID, err := b.lp.Send(ctx, b.roomID, content)
	if err != nil {
		b.log.Error("cannot send note to operators room", "error", err)
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
		b.log.Debug("sending note to customer room", "thread_id", threadID.String(), "note_id", int(note.ID))
		_, err = b.lp.Send(ctx, roomID, fullContent)
		if err != nil {
			b.log.Error("cannot send note to customer room", "error", err)
			b.SendNotice(ctx, b.roomID, linkpearl.UnwrapError(err).Error(), nil, linkpearl.RelatesTo(threadID))
			return
		}
	}
	b.log.Debug("marking note as synced", "thread_id", threadID.String(), "note_id", int(note.ID))
	if err := b.lp.SetAccountData(ctx, acID, map[string]string{"synced": "true"}); err != nil {
		b.log.Error("cannot set account data", "error", err)
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
		b.log.Warn("cannot parse file MXC URI", "error", err, "filename", fileName, "url", string(fileMXCURL))
		return nil
	}

	if fileContentURI.IsEmpty() {
		return nil
	}

	resp, err := b.lp.GetClient().Download(ctx, fileContentURI)
	if err != nil {
		b.log.Warn("cannot download file", "error", err, "filename", fileName, "url", string(fileMXCURL))
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
		b.log.Warn("cannot prepare file for decryption", "error", err)
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
	b.mu.Lock(key)
	defer b.mu.Unlock(key)

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
		b.log.Error("cannot update redmine issue", "error", updateErr)
	}
}

func (b *Bot) closeIssue(ctx context.Context, roomID id.RoomID, threadID id.EventID, text string) {
	key := "redmine_" + threadID.String()
	b.mu.Lock(key)
	defer b.mu.Unlock(key)

	issueIDStr, err := b.getRedmineMapping(ctx, threadID.String())
	if err != nil {
		return
	}
	issueID, err := strconv.Atoi(issueIDStr)
	if err != nil || issueID == 0 {
		return
	}
	log := b.log.With("issue_id", issueID)
	issue, err := b.redmine.GetIssue(int64(issueID), "attachments")
	if err != nil {
		log.Warn("cannot get redmine issue", "error", err)
		return
	}
	if issue.Attachments != nil {
		//nolint:gocritic // we can't do anything with it
		for _, attachment := range *issue.Attachments {
			if err := b.redmine.DeleteAttachment(attachment.ID); err != nil {
				log.Warn("cannot delete attachment", "error", err, "attachment_id", int(attachment.ID))
			}
		}
	}
	if updateErr := b.redmine.UpdateIssue(int64(issueID), b.redmine.StatusToID(redmine.Done), text); updateErr != nil {
		log.Warn("cannot close redmine issue", "error", updateErr)
		return
	}
	b.removeRedmineMapping(ctx, threadID.String())
	b.removeRedmineMapping(ctx, roomID.String())
	b.removeRedmineMapping(ctx, issueIDStr)
}
