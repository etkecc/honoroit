package redmine

import (
	"io"
	"net/http"
	"net/url"
	"strconv"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

// UploadRequest is a request to upload a file
// you can use either only Path to specify the file from the filesystem
// OR Stream to specify the file from a stream. In this case, Path is used as a filename
type UploadRequest struct {
	Path   string
	Stream io.Reader
}

// DeleteAttachment deletes an attachment by its ID
func (r *Redmine) DeleteAttachment(attachmentID int64) error {
	log := r.cfg.Log.With().Int64("attachment_id", attachmentID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring DeleteAttachment() call")
		return nil
	}
	if attachmentID == 0 {
		return nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := Retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.Del(nil, nil, url.URL{Path: "/attachments/" + strconv.FormatInt(attachmentID, 10) + ".json"}, http.StatusNoContent)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete attachment")
		return err
	}
	return nil
}

// uploadAttachments uploads attachments to Redmine, if any
// why it's not exported? Because Redmine REST API doesn't have a method to upload an attachment to an issue,
// attachment has to be uploaded first, and then attached to an issue. That can be done using NewIssue() or UpdateIssue() methods
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
			r.cfg.Log.Error().Err(err).Msg("failed to upload attachment")
			continue
		}
		if uploads == nil {
			uploads = &[]redmine.AttachmentUploadObject{}
		}
		*uploads = append(*uploads, upload)
	}
	return uploads
}
