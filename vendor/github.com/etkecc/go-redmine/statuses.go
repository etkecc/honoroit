package redmine

import redmine "github.com/nixys/nxs-go-redmine/v5"

const (
	WaitingForOperator Status = iota
	WaitingForCustomer
	Done
)

// Status is a type for issue statuses
// it should be used instead of sending raw status IDs
// to allow dynamic (re-)configuration
type Status int

// GetStatus returns the status of an issue
func (r *Redmine) GetStatus(issueID int64) (redmine.IssueStatusObject, error) {
	issue, err := r.GetIssue(issueID)
	if err != nil {
		return redmine.IssueStatusObject{}, err
	}
	return issue.Status, nil
}

// StatusToID converts a Status-typed number to actual status ID
func (r *Redmine) StatusToID(status Status) int64 {
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
		r.cfg.Log.Warn().Int("status", int(status)).Msg("unknown status")
	}

	return statusID
}
