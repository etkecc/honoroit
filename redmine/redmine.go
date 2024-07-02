package redmine

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

const (
	StatusNew = iota
	StatusInProgress
	StatusDone
)

type Redmine struct {
	api                *redmine.Context
	userID             int64
	projectID          int64
	trackerID          int64
	newStatusID        int64
	inProgressStatusID int64
	doneStatusID       int64
}

// New creates a new Redmine client
func New(host, apikey, projectIdentifier string, trackerID, newStatusID, inProgressStatusID, doneStatusID int) (*Redmine, error) {
	if host == "" || apikey == "" {
		return &Redmine{}, nil
	}

	r := &Redmine{
		api: redmine.Init(redmine.Settings{
			Endpoint: host,
			APIKey:   apikey,
		}),
		trackerID:          int64(trackerID),
		newStatusID:        int64(newStatusID),
		inProgressStatusID: int64(inProgressStatusID),
		doneStatusID:       int64(doneStatusID),
	}
	project, _, err := r.api.ProjectSingleGet(projectIdentifier, redmine.ProjectSingleGetRequest{})
	if err != nil {
		return nil, err
	}
	r.projectID = project.ID

	user, _, err := r.api.UserCurrentGet(redmine.UserCurrentGetRequest{})
	if err != nil {
		return nil, err
	}
	r.userID = user.ID
	return r, nil
}

// Enabled returns true if the Redmine client is enabled
func (r *Redmine) Enabled() bool {
	return r.api != nil
}

// NewIssue creates a new issue in Redmine
func (r *Redmine) NewIssue(subject, senderMedium, senderAddress, text string) (int64, error) {
	if !r.Enabled() {
		return 0, nil
	}
	if subject == "" || senderMedium == "" || senderAddress == "" || text == "" {
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
		return 0, err
	}
	return issue.ID, nil
}

func (r *Redmine) UpdateIssue(issueID int64, status int, text string) error {
	if !r.Enabled() {
		return nil
	}
	if issueID == 0 || text == "" {
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
		return fmt.Errorf("unknown status: %d", status)
	}

	statusCode, err := r.api.IssueUpdate(issueID, redmine.IssueUpdate{
		Issue: redmine.IssueUpdateObject{
			ProjectID: redmine.Int64Ptr(r.projectID),
			StatusID:  redmine.Int64Ptr(statusID),
			Notes:     redmine.StringPtr(text),
		},
	})
	if statusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return err
}

// IsClosed returns true if the issue is closed
func (r *Redmine) IsClosed(issueID int64) (bool, error) {
	if !r.Enabled() {
		return false, nil
	}
	if issueID == 0 {
		return false, nil
	}

	issue, statusCode, err := r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{})
	if statusCode == http.StatusNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return issue.Status.IsClosed || issue.Status.ID == r.doneStatusID, nil
}

// GetNotes returns the notes of an issue
func (r *Redmine) GetNotes(issueID int64) ([]*redmine.IssueJournalObject, error) {
	if !r.Enabled() {
		return nil, nil
	}
	if issueID == 0 {
		return nil, nil
	}

	issue, statusCode, err := r.api.IssueSingleGet(issueID, redmine.IssueSingleGetRequest{
		Includes: []redmine.IssueInclude{redmine.IssueIncludeJournals},
	})
	if statusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if issue.Journals == nil {
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
	return eligibleJournals, nil
}
