package redmine

import (
	"net/http"
	"net/url"
	"strconv"
)

/* Get */

type TimeEntryObject struct {
	ID        int64                `json:"id"`
	Project   IDName               `json:"project"`
	Issue     TimeEntryIssueObject `json:"issue"`
	User      IDName               `json:"user"`
	Activity  IDName               `json:"activity"`
	Hours     float64              `json:"hours"`
	Comments  string               `json:"comments"`
	SpentOn   string               `json:"spent_on"`
	CreatedOn string               `json:"created_on"`
	UpdatedOn string               `json:"updated_on"`
}

type TimeEntryIssueObject struct {
	ID int64 `json:"id"`
}

/* Create */

type TimeEntryCreate struct {
	TimeEntry TimeEntryCreateObject `json:"time_entry"`
}

type TimeEntryCreateObject struct {
	ProjectID  *string `json:"project_id,omitempty"`
	IssueID    *int64  `json:"issue_id,omitempty"`
	UserID     *int64  `json:"user_id,omitempty"`
	ActivityID int64   `json:"activity_id"`
	Hours      float64 `json:"hours"`
	Comments   string  `json:"comments"`
	SpentOn    *string `json:"spent_on,omitempty"`
}

/* Update */

type TimeEntryUpdate struct {
	TimeEntry TimeEntryUpdateObject `json:"time_entry"`
}

type TimeEntryUpdateObject struct {
	ProjectID  *string  `json:"project_id,omitempty"`
	IssueID    *int64   `json:"issue_id,omitempty"`
	UserID     *int64   `json:"user_id,omitempty"`
	ActivityID *int64   `json:"activity_id,omitempty"`
	Hours      *float64 `json:"hours,omitempty"`
	Comments   *string  `json:"comments,omitempty"`
	SpentOn    *string  `json:"spent_on,omitempty"`
}

/* Requests */

type TimeEntryAllGetRequest struct {
	Filters *TimeEntryGetRequestFilters
}

// Empty struct (uses as placeholder)
type TimeEntrySingleGetRequest struct {
}

type TimeEntryGetRequestFilters struct {
	userID      *int64
	projectID   *string
	spentOnFrom *string
	spentOnTo   *string
	activityID  *int64
}

/* Results */

// TimeEntryResult stores time entry requests processing result
type TimeEntryResult struct {
	TimeEntries []TimeEntryObject `json:"time_entries"`
	TotalCount  int64             `json:"total_count"`
	Offset      int64             `json:"offset"`
	Limit       int64             `json:"limit"`
}

/* Internal types */

type timeEntrySingleResult struct {
	TimeEntry TimeEntryObject `json:"time_entry"`
}

// TimeEntryAllGet gets info for all time entries satisfying specified filters
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries#Listing-time-entries
func (r *Context) TimeEntryAllGet(request TimeEntryAllGetRequest) (TimeEntryResult, StatusCode, error) {

	var (
		timeEntry TimeEntryResult
		offset    int64
		status    StatusCode
	)

	up := request.url()
	up.Set("limit", strconv.FormatInt(limitDefault, 10))

	for {

		var t TimeEntryResult

		up.Set("offset", strconv.FormatInt(offset, 10))

		s, err := r.Get(
			&t,
			url.URL{
				Path:     "/time_entries.json",
				RawQuery: up.Encode(),
			},
			http.StatusOK,
		)
		if err != nil {
			return timeEntry, s, err
		}

		status = s

		timeEntry.TimeEntries = append(timeEntry.TimeEntries, t.TimeEntries...)

		if offset+t.Limit >= t.TotalCount {
			timeEntry.TotalCount = t.TotalCount
			timeEntry.Limit = t.TotalCount

			break
		}

		offset += t.Limit
	}

	return timeEntry, status, nil
}

// TimeEntrySingleGet gets single time entry info by specific ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries#Showing-a-time-entry
func (r *Context) TimeEntrySingleGet(id int64, request TimeEntrySingleGetRequest) (TimeEntryObject, StatusCode, error) {

	var te timeEntrySingleResult

	status, err := r.Get(
		&te,
		url.URL{
			Path:     "/time_entries/" + strconv.FormatInt(id, 10) + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return te.TimeEntry, status, err
}

// TimeEntryCreate creates new time entry
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries#Creating-a-time-entry
func (r *Context) TimeEntryCreate(timeEntry TimeEntryCreate) (TimeEntryObject, StatusCode, error) {

	var te timeEntrySingleResult

	status, err := r.Post(
		timeEntry,
		&te,
		url.URL{
			Path: "/time_entries.json",
		},
		http.StatusCreated,
	)

	return te.TimeEntry, status, err
}

// TimeEntryUpdate updates time entry with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries#Updating-a-time-entry
func (r *Context) TimeEntryUpdate(id int64, timeEntry TimeEntryUpdate) (StatusCode, error) {

	status, err := r.Put(
		timeEntry,
		nil,
		url.URL{
			Path: "/time_entries/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// TimeEntryDelete deletes time entry with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries#Updating-a-time-entry
func (r *Context) TimeEntryDelete(id int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/time_entries/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

func (tr TimeEntryAllGetRequest) url() url.Values {

	v := url.Values{}

	if tr.Filters != nil {
		tr.Filters.url(&v)
	}

	return v
}

func (ur TimeEntrySingleGetRequest) url() url.Values {
	return url.Values{}
}

func TimeEntryGetRequestFiltersInit() *TimeEntryGetRequestFilters {
	return &TimeEntryGetRequestFilters{}
}

func (f *TimeEntryGetRequestFilters) ProjectSet(id string) *TimeEntryGetRequestFilters {
	f.projectID = &id
	return f
}

func (f *TimeEntryGetRequestFilters) SpentOnSet(from, to string) *TimeEntryGetRequestFilters {
	f.spentOnFrom = &from
	f.spentOnTo = &to
	return f
}

func (f *TimeEntryGetRequestFilters) UserIDSet(u int64) *TimeEntryGetRequestFilters {
	f.userID = &u
	return f
}

func (f *TimeEntryGetRequestFilters) ActivityIDSet(a int64) *TimeEntryGetRequestFilters {
	f.activityID = &a
	return f
}

func (f *TimeEntryGetRequestFilters) url(v *url.Values) {

	if f.projectID != nil {
		v.Set("project_id", *f.projectID)
	}

	if f.spentOnFrom != nil {
		v.Set("from", *f.spentOnFrom)
	}

	if f.spentOnTo != nil {
		v.Set("to", *f.spentOnTo)
	}

	if f.userID != nil {
		v.Set("user_id", strconv.FormatInt(*f.userID, 10))
	}

	if f.activityID != nil {
		v.Set("activity_id", strconv.FormatInt(*f.activityID, 10))
	}
}
