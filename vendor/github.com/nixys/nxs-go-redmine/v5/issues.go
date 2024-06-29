package redmine

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type IssueInclude string

const (
	IssueIncludeChildren        IssueInclude = "children" // used only: get single user
	IssueIncludeAttachments     IssueInclude = "attachments"
	IssueIncludeRelations       IssueInclude = "relations"
	IssueIncludeChangesets      IssueInclude = "changesets"       // used only: get single user
	IssueIncludeJournals        IssueInclude = "journals"         // used only: get single user
	IssueIncludeWatchers        IssueInclude = "watchers"         // used only: get single user
	IssueIncludeAllowedStatuses IssueInclude = "allowed_statuses" // used only: get single user
)

/* Get */

// IssueObject struct used for issues get operations
type IssueObject struct {
	ID                  int64                   `json:"id"`
	Project             IDName                  `json:"project"`
	Tracker             IDName                  `json:"tracker"`
	Status              IssueStatusObject       `json:"status"`
	Priority            IDName                  `json:"priority"`
	Author              IDName                  `json:"author"`
	AssignedTo          *IDName                 `json:"assigned_to"`
	Category            *IDName                 `json:"category"`
	FixedVersion        *IDName                 `json:"fixed_version"`
	Parent              *IssueParentObject      `json:"parent"`
	Subject             string                  `json:"subject"`
	Description         string                  `json:"description"`
	StartDate           *string                 `json:"start_date"`
	DueDate             *string                 `json:"due_date"`
	DoneRatio           int64                   `json:"done_ratio"`
	IsPrivate           int64                   `json:"is_private"`
	EstimatedHours      *float64                `json:"estimated_hours"`
	TotalEstimatedHours *float64                `json:"total_estimated_hours"`
	SpentHours          float64                 `json:"spent_hours"`
	TotalSpentHours     float64                 `json:"total_spent_hours"`
	CustomFields        []CustomFieldGetObject  `json:"custom_fields"`
	CreatedOn           string                  `json:"created_on"`
	UpdatedOn           string                  `json:"updated_on"`
	ClosedOn            string                  `json:"closed_on"`
	Children            *[]IssueChildrenObject  `json:"children"`         // used only: get single user and include specified
	Attachments         *[]AttachmentObject     `json:"attachments"`      // used only: include specified
	Relations           *[]IssueRelationObject  `json:"relations"`        // used only: include specified
	Changesets          *[]IssueChangesetObject `json:"changesets"`       // used only: get single user and include specified
	Journals            *[]IssueJournalObject   `json:"journals"`         // used only: get single user and include specified
	Watchers            *[]IDName               `json:"watchers"`         // used only: get single user and include specified
	AllowedStatuses     *[]IssueStatusObject    `json:"allowed_statuses"` // used only: get single user and include specified
}

// IssueParentObject struct used for issues get operations
type IssueParentObject struct {
	ID int64 `json:"id"`
}

// IssueChildrenObject struct used for issues get operations
type IssueChildrenObject struct {
	ID       int64                  `json:"id"`
	Tracker  IDName                 `json:"tracker"`
	Subject  string                 `json:"subject"`
	Children *[]IssueChildrenObject `json:"children"`
}

// IssueChangesetObject struct used for issues get operations
type IssueChangesetObject struct {
	Revision    string `json:"revision"`
	User        IDName `json:"user"`
	Comments    string `json:"comments"`
	CommittedOn string `json:"committed_on"`
}

// IssueRelationObject struct used for issues get operations
type IssueRelationObject struct {
	ID           int64  `json:"id"`
	IssueID      int64  `json:"issue_id"`
	IssueToID    int64  `json:"issue_to_id"`
	RelationType string `json:"relation_type"`
	Delay        *int64 `json:"delay"`
}

// IssueJournalObject struct used for issues get operations
type IssueJournalObject struct {
	ID           int64                      `json:"id"`
	User         IDName                     `json:"user"`
	Notes        string                     `json:"notes"`
	CreatedOn    string                     `json:"created_on"`
	PrivateNotes bool                       `json:"private_notes"`
	Details      []IssueJournalDetailObject `json:"details"`
}

// IssueJournalDetailObject struct used for issues get operations
type IssueJournalDetailObject struct {
	Property string `json:"property"`
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

/* Create */

// IssueCreate struct used for issues create operations
type IssueCreate struct {
	Issue IssueCreateObject `json:"issue"`
}

type IssueCreateObject struct {
	ProjectID      int64                      `json:"project_id"`
	TrackerID      *int64                     `json:"tracker_id,omitempty"`
	StatusID       *int64                     `json:"status_id,omitempty"`
	PriorityID     *int64                     `json:"priority_id,omitempty"`
	Subject        string                     `json:"subject"`
	Description    *string                    `json:"description,omitempty"`
	StartDate      *string                    `json:"start_date,omitempty"`
	DueDate        *string                    `json:"due_date,omitempty"`
	CategoryID     *int64                     `json:"category_id,omitempty"`
	FixedVersionID *int64                     `json:"fixed_version_id,omitempty"`
	AssignedToID   *int64                     `json:"assigned_to_id,omitempty"`
	ParentIssueID  *int64                     `json:"parent_issue_id,omitempty"`
	CustomFields   *[]CustomFieldUpdateObject `json:"custom_fields,omitempty"`
	WatcherUserIDs *[]int64                   `json:"watcher_user_ids,omitempty"`
	IsPrivate      *bool                      `json:"is_private,omitempty"`
	EstimatedHours *float64                   `json:"estimated_hours,omitempty"`
	Uploads        *[]AttachmentUploadObject  `json:"uploads,omitempty"`
}

/* Update */

// IssueUpdate struct used for issues update operations
type IssueUpdate struct {
	Issue IssueUpdateObject `json:"issue"`
}

type IssueUpdateObject struct {
	ProjectID      *int64                     `json:"project_id,omitempty"`
	TrackerID      *int64                     `json:"tracker_id,omitempty"`
	StatusID       *int64                     `json:"status_id,omitempty"`
	PriorityID     *int64                     `json:"priority_id,omitempty"`
	Subject        *string                    `json:"subject,omitempty"`
	Description    *string                    `json:"description,omitempty"`
	StartDate      *string                    `json:"start_date,omitempty"`
	DueDate        *string                    `json:"due_date,omitempty"`
	CategoryID     *int64                     `json:"category_id,omitempty"`
	FixedVersionID *int64                     `json:"fixed_version_id,omitempty"`
	AssignedToID   *int64                     `json:"assigned_to_id,omitempty"`
	ParentIssueID  *int64                     `json:"parent_issue_id,omitempty"`
	CustomFields   *[]CustomFieldUpdateObject `json:"custom_fields,omitempty"`
	IsPrivate      *bool                      `json:"is_private,omitempty"`
	EstimatedHours *float64                   `json:"estimated_hours,omitempty"`
	Uploads        *[]AttachmentUploadObject  `json:"uploads,omitempty"`
	Notes          *string                    `json:"notes,omitempty"`
	PrivateNotes   *bool                      `json:"private_notes,omitempty"`
}

/* Requests */

// IssueAllGetRequest contains data for making request to get all issues satisfying specified filters
type IssueAllGetRequest struct {
	Sort     *IssueGetRequestSort
	Includes []IssueInclude
	Filters  *IssueGetRequestFilters
}

// IssueMultiGetRequest contains data for making request to get limited issues count satisfying specified filters
type IssueMultiGetRequest struct {
	Sort     *IssueGetRequestSort
	Includes []IssueInclude
	Filters  *IssueGetRequestFilters
	Offset   int64
	Limit    int64
}

// IssueSingleGetRequest contains data for making request to get specified issue
type IssueSingleGetRequest struct {
	Includes []IssueInclude
}

type IssueGetRequestSort struct {
	field string
	desc  bool
}

// IssueGetRequestFilters contains data for making issues get request
type IssueGetRequestFilters struct {
	fields map[string][]string
	cf     map[int64]string
}

/* Results */

// IssueResult stores issues requests processing result
type IssueResult struct {
	Issues     []IssueObject `json:"issues"`
	TotalCount int64         `json:"total_count"`
	Offset     int64         `json:"offset"`
	Limit      int64         `json:"limit"`
}

/* Internal types */

type issueSingleResult struct {
	Issue IssueObject `json:"issue"`
}

type issueWatcherAdd struct {
	UserID int64 `json:"user_id"`
}

func (ii IssueInclude) String() string {
	return string(ii)
}

// IssuesAllGet gets info for all issues satisfying specified filters
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Listing-issues
func (r *Context) IssuesAllGet(request IssueAllGetRequest) (IssueResult, StatusCode, error) {

	var (
		issues IssueResult
		offset int64
		status StatusCode
	)

	up := request.url()
	up.Set("limit", strconv.FormatInt(limitDefault, 10))

	for {

		var i IssueResult

		up.Set("offset", strconv.FormatInt(offset, 10))

		s, err := r.Get(
			&i,
			url.URL{
				Path:     "/issues.json",
				RawQuery: request.url().Encode(),
			},
			http.StatusOK,
		)
		if err != nil {
			return issues, s, err
		}

		status = s

		issues.Issues = append(issues.Issues, i.Issues...)

		if offset+i.Limit >= i.TotalCount {
			issues.TotalCount = i.TotalCount
			issues.Limit = i.TotalCount

			break
		}

		offset += i.Limit
	}

	return issues, status, nil
}

// IssuesMultiGet gets info for multiple issues satisfying specified filters
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Listing-issues
func (r *Context) IssuesMultiGet(request IssueMultiGetRequest) (IssueResult, StatusCode, error) {

	var i IssueResult

	s, err := r.Get(
		&i,
		url.URL{
			Path:     "/issues.json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return i, s, err
}

// IssueSingleGet gets single issue info
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Showing-an-issue
func (r *Context) IssueSingleGet(id int64, request IssueSingleGetRequest) (IssueObject, StatusCode, error) {

	var i issueSingleResult

	status, err := r.Get(
		&i,
		url.URL{
			Path:     "/issues/" + strconv.FormatInt(id, 10) + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return i.Issue, status, err
}

// IssueCreate creates new issue
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Creating-an-issue
func (r *Context) IssueCreate(issue IssueCreate) (IssueObject, StatusCode, error) {

	var i issueSingleResult

	status, err := r.Post(
		issue, &i,
		url.URL{
			Path: "/issues.json",
		},
		http.StatusCreated,
	)

	return i.Issue, status, err
}

// IssueUpdate updates issue with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Projects#Updating-a-project
func (r *Context) IssueUpdate(id int64, issue IssueUpdate) (StatusCode, error) {

	status, err := r.Put(
		issue,
		nil,
		url.URL{
			Path: "/issues/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// IssueDelete deletes issue with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Deleting-an-issue
func (r *Context) IssueDelete(id int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/issues/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// IssueWatcherAdd adds watcher into issue with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Adding-a-watcher
func (r *Context) IssueWatcherAdd(id int64, userID int64) (StatusCode, error) {

	status, err := r.Post(
		issueWatcherAdd{
			UserID: userID,
		},
		nil,
		url.URL{
			Path: "/issues/" + strconv.FormatInt(id, 10) + "/watchers.json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// IssueWatcherDelete deletes watcher from issue with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Issues#Removing-a-watcher
func (r *Context) IssueWatcherDelete(id int64, userID int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/issues/" + strconv.FormatInt(id, 10) + "/watchers/" + strconv.FormatInt(userID, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

func (ir IssueAllGetRequest) url() url.Values {

	v := url.Values{}

	if ir.Sort != nil {
		ir.Sort.url(&v)
	}

	if len(ir.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range ir.Includes {
						is = append(is, i.String())
					}
					return is
				}(),
				",",
			),
		)
	}

	if ir.Filters != nil {
		ir.Filters.url(&v)
	}

	return v
}

func (ir IssueMultiGetRequest) url() url.Values {

	v := url.Values{}

	if ir.Sort != nil {
		ir.Sort.url(&v)
	}

	if len(ir.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range ir.Includes {
						is = append(is, i.String())
					}
					return is
				}(),
				",",
			),
		)
	}

	if ir.Filters != nil {
		ir.Filters.url(&v)
	}

	v.Set("offset", strconv.FormatInt(ir.Offset, 10))
	v.Set("limit", strconv.FormatInt(ir.Limit, 10))

	return v
}

func (ir IssueSingleGetRequest) url() url.Values {

	v := url.Values{}

	if len(ir.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range ir.Includes {
						is = append(is, i.String())
					}
					return is
				}(),
				",",
			),
		)
	}

	return v
}

func IssueGetRequestFiltersInit() *IssueGetRequestFilters {
	return &IssueGetRequestFilters{
		fields: make(map[string][]string),
		cf:     make(map[int64]string),
	}
}

func (f *IssueGetRequestFilters) FieldAdd(field string, values ...string) *IssueGetRequestFilters {
	f.fields[field] = append([]string{}, values...)
	return f
}

func (f *IssueGetRequestFilters) CustomFieldAdd(id int64, value string) *IssueGetRequestFilters {
	f.cf[id] = value
	return f
}

func (f *IssueGetRequestFilters) url(v *url.Values) {

	// Filter fields (e.g. `issue_id`, `tracker_id`, etc)
	for n, s := range f.fields {
		v.Set(n, strings.Join(s, ","))
	}

	// Custom fields
	for id, value := range f.cf {
		v.Set("cf_"+strconv.FormatInt(id, 10), value)
	}
}

func IssueGetRequestSortInit() *IssueGetRequestSort {
	return &IssueGetRequestSort{}
}

func (s *IssueGetRequestSort) Set(field string, desc bool) *IssueGetRequestSort {
	s.field = field
	s.desc = desc
	return s
}

func (s *IssueGetRequestSort) url(v *url.Values) {

	f := s.field
	if s.desc == true {
		f += ":desc"
	}

	v.Set("sort", f)
}
