package redmine

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	redmine "github.com/nixys/nxs-go-redmine/v5"
	"github.com/rs/zerolog"
)

const (
	StatusNew = iota
	StatusInProgress
	StatusDone
)

type Redmine struct {
	log                *zerolog.Logger
	api                *redmine.Context
	userID             int64
	projectID          int64
	trackerID          int64
	newStatusID        int64
	inProgressStatusID int64
	doneStatusID       int64
}

// New creates a new Redmine client
func New(log *zerolog.Logger, host, apikey, projectIdentifier string, trackerID, newStatusID, inProgressStatusID, doneStatusID int) (*Redmine, error) {
	empty := &Redmine{log: log}
	if host == "" || apikey == "" {
		return empty, nil
	}

	r := &Redmine{
		api: redmine.Init(redmine.Settings{
			Endpoint: host,
			APIKey:   apikey,
		}),
		log:                log,
		trackerID:          int64(trackerID),
		newStatusID:        int64(newStatusID),
		inProgressStatusID: int64(inProgressStatusID),
		doneStatusID:       int64(doneStatusID),
	}
	project, _, err := r.api.ProjectSingleGet(projectIdentifier, redmine.ProjectSingleGetRequest{})
	if err != nil {
		return empty, err
	}
	r.projectID = project.ID

	user, _, err := r.api.UserCurrentGet(redmine.UserCurrentGetRequest{})
	if err != nil {
		return empty, err
	}
	r.userID = user.ID
	return r, nil
}

// Enabled returns true if the Redmine client is enabled
func (r *Redmine) Enabled() bool {
	return r.api != nil
}

// NewIssue creates a new issue in Redmine
func (r *Redmine) NewIssue(threadID, subject, senderMedium, senderAddress, text string) (int64, error) {
	if !r.Enabled() {
		r.log.Debug().Msg("redmine is disabled, ignoring NewIssue() call")
		return 0, nil
	}
	if subject == "" || senderMedium == "" || senderAddress == "" || text == "" {
		r.log.Warn().Str("thread_id", threadID).Msg("missing required fields, ignoring NewIssue() call")
		return 0, nil
	}

	var description strings.Builder
	description.WriteString(fmt.Sprintf("Sender: `%s` (%s)\n\n", senderAddress, senderMedium))
	description.WriteString(text)
	text = description.String()
	issue, _, err := r.api.IssueCreate(redmine.IssueCreate{
		Issue: redmine.IssueCreateObject{
			ProjectID:   r.projectID,
			TrackerID:   redmine.Int64Ptr(r.trackerID),
			StatusID:    redmine.Int64Ptr(r.newStatusID),
			Subject:     subject,
			Description: redmine.StringPtr(text),
		},
	})
	if err != nil {
		r.log.Error().Err(err).Str("thread_id", threadID).Msg("failed to create issue")
		return 0, err
	}
	r.log.Info().Str("thread_id", threadID).Int64("issue_id", issue.ID).Msg("issue created")
	return issue.ID, nil
}

// UpdateIssue updates the status and notes of an issue
func (r *Redmine) UpdateIssue(issueID int64, status int, text string) error {
	if !r.Enabled() {
		r.log.Debug().Msg("redmine is disabled, ignoring UpdateIssue() call")
		return nil
	}
	if issueID == 0 || text == "" {
		r.log.Debug().Msg("missing required fields, ignoring UpdateIssue() call")
		return nil
	}

	var statusID int64
	switch status {
	case StatusNew:
		statusID = r.newStatusID
	case StatusInProgress:
		statusID = r.inProgressStatusID
	case StatusDone:
		statusID = r.doneStatusID
	default:
		r.log.Error().Int("status", status).Msg("unknown status")
		return fmt.Errorf("unknown status: %d", status)
	}

	maxRetries := 5
	delay := 5 * time.Second
	for i := 1; i <= maxRetries; i++ {
		statusCode, err := r.api.IssueUpdate(issueID, redmine.IssueUpdate{
			Issue: redmine.IssueUpdateObject{
				ProjectID: redmine.Int64Ptr(r.projectID),
				StatusID:  redmine.Int64Ptr(statusID),
				Notes:     redmine.StringPtr(text),
			},
		})
		if statusCode == http.StatusNotFound {
			r.log.Warn().Int64("issue_id", issueID).Msg("issue not found")
			return nil
		}
		if statusCode > 499 {
			if i < maxRetries {
				r.log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Int64("issue_id", issueID).Msg("failed to update issue, retrying")
				time.Sleep(delay * time.Duration(i))
				continue
			}
			r.log.Warn().Int("retries", i).Int("status_code", int(statusCode)).Err(err).Int64("issue_id", issueID).Msg("failed to update issue, giving up")
		}

		if err != nil {
			r.log.Error().Err(err).Int64("issue_id", issueID).Msg("failed to update issue")
			return err
		}
	}
	return nil
}

// IsClosed returns true if the issue is closed
func (r *Redmine) IsClosed(issueID int64) (bool, error) {
	if !r.Enabled() {
		r.log.Debug().Msg("redmine is disabled, ignoring IsClosed() call")
		return false, nil
	}
	if issueID == 0 {
		return false, nil
	}

	issue, statusCode, err := r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
	if statusCode == http.StatusNotFound {
		r.log.Warn().Int64("issue_id", issueID).Msg("issue not found")
		return false, nil
	}
	if err != nil {
		r.log.Error().Err(err).Int64("issue_id", issueID).Msg("failed to get issue")
		return false, err
	}
	return issue.Status.IsClosed || issue.Status.ID == r.doneStatusID, nil
}

// GetNotes returns the notes of an issue
func (r *Redmine) GetNotes(issueID int64) ([]*redmine.IssueJournalObject, error) {
	if !r.Enabled() {
		r.log.Debug().Msg("redmine is disabled, ignoring GetNotes() call")
		return nil, nil
	}
	if issueID == 0 {
		return nil, nil
	}

	issue, statusCode, err := r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{
		Includes: []redmine.IssueInclude{redmine.IssueIncludeJournals},
	})
	if statusCode == http.StatusNotFound {
		r.log.Warn().Int64("issue_id", issueID).Msg("issue not found")
		return nil, nil
	}
	if err != nil {
		r.log.Error().Err(err).Int64("issue_id", issueID).Msg("failed to get issue")
		return nil, err
	}
	if issue.Journals == nil {
		r.log.Debug().Int64("issue_id", issueID).Msg("no journals found")
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
		if journal.User.ID == r.userID {
			continue
		}
		if journal.Notes == "" {
			continue
		}
		eligibleJournals = append(eligibleJournals, &journal)
	}
	r.log.Debug().Int("journals", len(eligibleJournals)).Int64("issue_id", issueID).Msg("journals found")
	return eligibleJournals, nil
}
