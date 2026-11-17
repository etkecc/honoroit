package redmine

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

// UploadRequest uploads a file: set Path for a filesystem file, or Stream (Path becomes the filename).
type UploadRequest struct {
	Path   string
	Stream io.Reader
}

// DeleteAttachment deletes an attachment by its ID
func (r *Redmine) DeleteAttachment(attachmentID int64) error {
	log := r.cfg.Log.With(slog.Int64("attachment_id", attachmentID))
	if !r.Enabled() {
		log.Debug("redmine is disabled, ignoring DeleteAttachment() call")
		return nil
	}
	if attachmentID == 0 {
		return nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := Retry(log, func() (redmine.StatusCode, error) {
		return r.cfg.api.Del(nil, nil, url.URL{Path: "/attachments/" + strconv.FormatInt(attachmentID, 10) + ".json"}, http.StatusNoContent)
	})
	if err != nil {
		log.Error("failed to delete attachment", slog.Any("error", err))
		return err
	}
	return nil
}

// uploadAttachments uploads to Redmine; unexported: Redmine needs upload-then-attach via NewIssue/UpdateIssue.
func (r *Redmine) uploadAttachments(files ...*UploadRequest) *[]redmine.AttachmentUploadObject {
	var uploads *[]redmine.AttachmentUploadObject
	for _, req := range files {
		if req == nil {
			continue
		}
		upload, err := RetryResult(r.cfg.Log, func() (redmine.AttachmentUploadObject, redmine.StatusCode, error) {
			if req.Stream == nil {
				return r.cfg.api.AttachmentUpload(req.Path)
			}
			if streamCloser, ok := req.Stream.(io.Closer); ok {
				defer streamCloser.Close()
			}
			return r.cfg.api.AttachmentUploadStream(req.Stream, req.Path)
		})
		if err != nil {
			r.cfg.Log.Error("failed to upload attachment", slog.Any("error", err))
			continue
		}
		if uploads == nil {
			uploads = &[]redmine.AttachmentUploadObject{}
		}
		*uploads = append(*uploads, upload)
	}
	return uploads
}
