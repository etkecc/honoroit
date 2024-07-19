package redmine

import (
	"fmt"
	"sort"
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
}

// Redmine is a Redmine client
type Redmine struct {
	wg  sync.WaitGroup
	cfg *Config
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

func (r *Redmine) Configure(options ...Option) {
	r.cfg.apply(options...)
}

// NewIssue creates a new issue in Redmine
func (r *Redmine) NewIssue(subject, senderMedium, senderAddress, text string) (int64, error) {
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
func (r *Redmine) UpdateIssue(issueID int64, status Status, text string) error {
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
		log.Error().Int("status", int(status)).Msg("unknown status")
		return fmt.Errorf("unknown status: %d", status)
	}

	r.wg.Add(1)
	defer r.wg.Done()

	err := retry(&log, func() (redmine.StatusCode, error) {
		return r.cfg.api.IssueUpdate(issueID, redmine.IssueUpdate{
			Issue: redmine.IssueUpdateObject{
				ProjectID: redmine.Int64Ptr(r.cfg.ProjectID),
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
	log := r.cfg.Log.With().Int64("issue_id", issueID).Logger()
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
		return r.cfg.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to get issue")
		return false, err
	}
	return issue.Status.IsClosed || issue.Status.ID == r.cfg.DoneStatusID, nil
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

// Shutdown waits for all goroutines to finish
func (r *Redmine) Shutdown() {
	r.wg.Wait()
}
