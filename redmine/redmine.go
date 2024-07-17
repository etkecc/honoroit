package redmine

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	redmine "github.com/nixys/nxs-go-redmine/v5"
	"github.com/rs/zerolog"
)

const (
	StatusNew = iota
	StatusInProgress
	StatusDone
)

type Redmine struct {
	wg                 sync.WaitGroup
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
	log := r.log.With().Str("thread_id", threadID).Logger()
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
		return r.api.IssueCreate(redmine.IssueCreate{
			Issue: redmine.IssueCreateObject{
				ProjectID:   r.projectID,
				TrackerID:   redmine.Int64Ptr(r.trackerID),
				StatusID:    redmine.Int64Ptr(r.newStatusID),
				Subject:     subject,
				Description: redmine.StringPtr(text),
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

// UpdateIssue updates the status and notes of an issue
func (r *Redmine) UpdateIssue(issueID int64, status int, text string) error {
	log := r.log.With().Int64("issue_id", issueID).Logger()
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
	case StatusNew:
		statusID = r.newStatusID
	case StatusInProgress:
		statusID = r.inProgressStatusID
	case StatusDone:
		statusID = r.doneStatusID
	default:
		log.Error().Int("status", status).Msg("unknown status")
		return fmt.Errorf("unknown status: %d", status)
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.api.IssueUpdate(issueID, redmine.IssueUpdate{
			Issue: redmine.IssueUpdateObject{
				ProjectID: redmine.Int64Ptr(r.projectID),
				StatusID:  redmine.Int64Ptr(statusID),
				Notes:     redmine.StringPtr(text),
			},
		})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to update issue")
		return err
	}
	return nil
}

// IsClosed returns true if the issue is closed
func (r *Redmine) IsClosed(issueID int64) (bool, error) {
	log := r.log.With().Int64("issue_id", issueID).Logger()
	if !r.Enabled() {
		log.Debug().Msg("redmine is disabled, ignoring IsClosed() call")
		return false, nil
	}
	if issueID == 0 {
		return false, nil
	}

	r.wg.Add(1)
	defer r.wg.Done()

	issue, err := retryResult(&log, func() (redmine.IssueObject, redmine.StatusCode, error) {
		return r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return false, err
	}
	return issue.Status.IsClosed || issue.Status.ID == r.doneStatusID, nil
}

// GetNotes returns the notes of an issue
func (r *Redmine) GetNotes(issueID int64) ([]*redmine.IssueJournalObject, error) {
	log := r.log.With().Int64("issue_id", issueID).Logger()
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
		return r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{
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
		if journal.User.ID == r.userID {
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

// Shutdown waits for all goroutines to finish
func (r *Redmine) Shutdown() {
	r.wg.Wait()
}
