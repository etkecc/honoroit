package redmine

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

const (
	WaitingForOperator Status = iota
	WaitingForCustomer
	Done
)

// Status is a type for issue statuses
// it should be used instead of sending raw status IDs
// to allow dynamic (re-)configuration
type Status int

// API is an interface for Redmine API
type API interface {
	ProjectSingleGet(identifier string, req redmine.ProjectSingleGetRequest) (redmine.ProjectObject, redmine.StatusCode, error)
	UserCurrentGet(req redmine.UserCurrentGetRequest) (redmine.UserObject, redmine.StatusCode, error)
	IssueCreate(req redmine.IssueCreate) (redmine.IssueObject, redmine.StatusCode, error)
	IssueUpdate(id int64, req redmine.IssueUpdate) (redmine.StatusCode, error)
	IssueSingleGet(id int64, req redmine.IssueSingleGetRequest) (redmine.IssueObject, redmine.StatusCode, error)
	IssueDelete(id int64) (redmine.StatusCode, error)
	AttachmentUpload(filePath string) (redmine.AttachmentUploadObject, redmine.StatusCode, error)
	AttachmentUploadStream(f io.Reader, fileName string) (redmine.AttachmentUploadObject, redmine.StatusCode, error)
	Del(in, out any, uri url.URL, statusExpected redmine.StatusCode) (redmine.StatusCode, error)
	Post(in, out any, uri url.URL, statusExpected redmine.StatusCode) (redmine.StatusCode, error)
}

// IssueRelation is a minimal struct for issue relations
type IssueRelation struct {
	IssueToID    int64  `json:"issue_to_id"`
	RelationType string `json:"relation_type"`
}

// IssueRelationRequest is a request to create a new issue relation
type IssueRelationRequest struct {
	Relation IssueRelation `json:"relation"`
}

// Redmine is a Redmine client
type Redmine struct {
	wg  sync.WaitGroup
	cfg *Config
}

// UploadRequest is a request to upload a file
// you can use either only Path to specify the file from the filesystem
// OR Stream to specify the file from a stream. In this case, Path is used as a filename
type UploadRequest struct {
	Path   string
	Stream io.Reader
}

// New creates a new Redmine client
func New(options ...Option) (*Redmine, error) {
	cfg := NewConfig(options...)
	r := &Redmine{cfg: cfg}
	if !cfg.Enabled() {
		return r, nil
	}

	if cfg.ProjectID == 0 {
		project, _, err := r.cfg.api.ProjectSingleGet(cfg.ProjectIdentifier, redmine.ProjectSingleGetRequest{})
		if err != nil {
			return r, err
		}
		r.cfg.ProjectID = project.ID
	}

	if cfg.UserID == 0 {
		user, _, err := r.cfg.api.UserCurrentGet(redmine.UserCurrentGetRequest{})
		if err != nil {
			return r, err
		}
		r.cfg.UserID = user.ID
	}
	return r, nil
}

// Enabled returns true if the Redmine client is enabled
func (r *Redmine) Enabled() bool {
	return r.cfg.Enabled()
}

// Configure applies the new configuration options
func (r *Redmine) Configure(options ...Option) {
	r.cfg.apply(options...)
}

// uploadAttachments uploads attachments to Redmine, if any
func (r *Redmine) uploadAttachments(files ...*UploadRequest) *[]redmine.AttachmentUploadObject {
	var uploads *[]redmine.AttachmentUploadObject
	for _, req := range files {
		if req == nil {
			r.cfg.Log.Warn().Msg("nil upload request")
		}
		upload, err := retryResult(r.cfg.Log, func() (redmine.AttachmentUploadObject, redmine.StatusCode, error) {
			if req.Stream == nil {
				return r.cfg.api.AttachmentUpload(req.Path)
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

// NewIssue creates a new issue in Redmine
func (r *Redmine) NewIssue(subject, senderMedium, senderAddress, text string, files ...*UploadRequest) (int64, error) {
	log := r.cfg.Log.With().Str(senderMedium, senderAddress).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring NewIssue() call")
		return 0, nil
	}
	if subject == "" || senderMedium == "" || senderAddress == "" || text == "" {
		log.Warn().Msg("missing required fields, ignoring NewIssue() call")
		return 0, nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	var description strings.Builder
	description.WriteString(fmt.Sprintf("Sender: `%s` (%s)\n\n", senderAddress, senderMedium))
	description.WriteString(text)
	text = description.String()
	issue, err := retryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
		return r.cfg.api.IssueCreate(redmine.IssueCreate{
			Issue: redmine.IssueCreateObject{
				ProjectID:   r.cfg.ProjectID,
				TrackerID:   redmine.Int64Ptr(r.cfg.TrackerID),
				StatusID:    redmine.Int64Ptr(r.cfg.WaitingForOperatorStatusID),
				Subject:     subject,
				Description: redmine.StringPtr(text),
				Uploads:     r.uploadAttachments(files...),
			},
		})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create issue")
		return 0, err
	}
	log.Info().Int64("issue_id", issue.ID).Msg("issue created")
	return issue.ID, nil
}

// GetIssue returns an issue by its ID
func (r *Redmine) GetIssue(issueID int64, includes ...redmine.IssueInclude) (redmine.IssueObject, error) {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	var issue redmine.IssueObject
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring GetIssue() call")
		return issue, nil
	}
	if issueID == 0 {
		return issue, nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	issue, err := retryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
		return r.cfg.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{Includes: includes})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return issue, err
	}
	return issue, nil
}

// UpdateIssue updates the status using one of the constants and notes of an issue
func (r *Redmine) UpdateIssue(issueID int64, status Status, text string, files ...*UploadRequest) error {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring UpdateIssue() call")
		return nil
	}
	if issueID == 0 || text == "" {
		log.Debug().Msg("missing required fields, ignoring UpdateIssue() call")
		return nil
	}

	var statusID int64
	switch status {
	case WaitingForOperator:
		statusID = r.cfg.WaitingForOperatorStatusID
	case WaitingForCustomer:
		statusID = r.cfg.WaitingForCustomerStatusID
	case Done:
		statusID = r.cfg.DoneStatusID
	default:
		statusID = int64(status)
		log.Warn().Int("status", int(status)).Msg("unknown status")
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.IssueUpdate(issueID, redmine.IssueUpdate{
			Issue: redmine.IssueUpdateObject{
				ProjectID: redmine.Int64Ptr(r.cfg.ProjectID),
				StatusID:  redmine.Int64Ptr(statusID),
				Notes:     redmine.StringPtr(text),
				Uploads:   r.uploadAttachments(files...),
			},
		})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update issue")
		return err
	}
	return nil
}

// DeleteIssue deletes an issue by its ID
func (r *Redmine) DeleteIssue(issueID int64) error {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring DeleteIssue() call")
		return nil
	}
	if issueID == 0 {
		return nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.IssueDelete(issueID)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete issue")
		return err
	}
	return nil
}

// GetStatus returns the status of an issue
func (r *Redmine) GetStatus(issueID int64) (redmine.IssueStatusObject, error) {
	issue, err := r.GetIssue(issueID)
	if err != nil {
		return redmine.IssueStatusObject{}, err
	}
	return issue.Status, nil
}

// IsClosed returns true if the issue is closed or has "Done" status
func (r *Redmine) IsClosed(issueID int64) (bool, error) {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	status, err := r.GetStatus(issueID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return false, err
	}

	if status.ID == 0 {
		return false, nil
	}

	return status.IsClosed || status.ID == r.cfg.DoneStatusID, nil
}

// GetNotes returns the notes of an issue
func (r *Redmine) GetNotes(issueID int64) ([]*redmine.IssueJournalObject, error) {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring GetNotes() call")
		return nil, nil
	}
	if issueID == 0 {
		return nil, nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	issue, err := retryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
		return r.cfg.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{
			Includes: []redmine.IssueInclude{redmine.IssueIncludeJournals},
		})
	})
	if issue.ID == 0 {
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return nil, err
	}
	if issue.Journals == nil {
		log.Debug().Msg("no journals found")
		return nil, nil
	}
	journals := *issue.Journals
	// sort journals by id asc
	sort.Slice(journals, func(i, j int) bool {
		return journals[i].ID < journals[j].ID
	})
	eligibleJournals := []*redmine.IssueJournalObject{}
	for _, journal := range journals {
		journal := journal
		if journal.User.ID == r.cfg.UserID {
			continue
		}
		if journal.Notes == "" {
			continue
		}
		eligibleJournals = append(eligibleJournals, &journal)
	}
	log.Debug().Int("journals", len(eligibleJournals)).Msg("journals found")
	return eligibleJournals, nil
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

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.Del(nil, nil, url.URL{Path: "/attachments/" + strconv.FormatInt(attachmentID, 10) + ".json"}, http.StatusNoContent)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete attachment")
		return err
	}
	return nil
}

func (r *Redmine) NewIssueRelation(issueID, relatedIssueID int64, relationType string) error {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Int64("related_issue_id", relatedIssueID).Str("relation_type", relationType).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring NewIssueRelation() call")
		return nil
	}
	if issueID == 0 || relatedIssueID == 0 {
		log.Warn().Msg("missing required fields, ignoring NewIssueRelation() call")
		return nil
	}
	if relationType == "" {
		relationType = "relates"
	}

	r.wg.Add(1)
	defer r.wg.Done()

	relation := &IssueRelationRequest{
		Relation: IssueRelation{
			IssueToID:    relatedIssueID,
			RelationType: relationType,
		},
	}

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.Post(relation, nil, url.URL{Path: fmt.Sprintf("/issues/%d/relations.json", issueID)}, http.StatusCreated)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create issue relation")
		return err
	}
	return nil
}

// Shutdown waits for all goroutines to finish
func (r *Redmine) Shutdown() {
	r.wg.Wait()
}
