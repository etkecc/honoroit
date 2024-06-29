package redmine

import (
	"net/http"
	"net/url"
)

/* Get */

// IssueStatusObject struct used for issue_statuses get operations
type IssueStatusObject struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	IsClosed bool   `json:"is_closed"`
}

/* Internal types */

type issueStatusAllResult struct {
	IssueStatuses []IssueStatusObject `json:"issue_statuses"`
}

// IssueStatusAllGet gets info for all issue statuses
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_IssueStatuses#GET
func (r *Context) IssueStatusAllGet() ([]IssueStatusObject, StatusCode, error) {

	var i issueStatusAllResult

	status, err := r.Get(
		&i,
		url.URL{
			Path: "/issue_statuses.json",
		},
		http.StatusOK,
	)

	return i.IssueStatuses, status, err
}
