package redmine

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

// IssueRelation is a minimal struct for issue relations
type IssueRelation struct {
	IssueToID    int64  `json:"issue_to_id"`
	RelationType string `json:"relation_type"`
}

// IssueRelationRequest is a request to create a new issue relation
type IssueRelationRequest struct {
	Relation IssueRelation `json:"relation"`
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

	issue, err := RetryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
		return r.cfg.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{Includes: includes})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return issue, err
	}
	return issue, nil
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
	issue, err := RetryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
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

// UpdateIssue updates the status using one of the constants and notes of an issue
func (r *Redmine) UpdateIssue(issueID, statusID int64, text string, files ...*UploadRequest) error {
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring UpdateIssue() call")
		return nil
	}
	if issueID == 0 || text == "" {
		log.Debug().Msg("missing required fields, ignoring UpdateIssue() call")
		return nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := Retry(&log, func() (redmine.StatusCode, error) {
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

	err := Retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.IssueDelete(issueID)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete issue")
		return err
	}
	return nil
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

	issue, err := RetryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
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

	err := Retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.Post(relation, nil, url.URL{Path: fmt.Sprintf("/issues/%d/relations.json", issueID)}, http.StatusCreated)
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create issue relation")
		return err
	}
	return nil
}
