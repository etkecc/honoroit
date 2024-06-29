package redmine

import (
	"net/http"
	"net/url"
)

/* Get */

// EnumerationPriorityObject struct used for priorities get operations
type EnumerationPriorityObject struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Active    bool   `json:"active"`
}

// EnumerationTimeEntryActivityObject struct used for time entry activities get operations
type EnumerationTimeEntryActivityObject struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Active    bool   `json:"active"`
}

// EnumerationDocumentCategoryObject struct used for document categories get operations
type EnumerationDocumentCategoryObject struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Active    bool   `json:"active"`
}

/* Internal types */

type enumerationPrioritiesAllResult struct {
	Priorities []EnumerationPriorityObject `json:"issue_priorities"`
}

type enumerationTimeEntryActivitiesAllResult struct {
	TimeEntryActivities []EnumerationTimeEntryActivityObject `json:"time_entry_activities"`
}

type enumerationDocumentCategoriesAllResult struct {
	DocumentCategories []EnumerationDocumentCategoryObject `json:"document_categories"`
}

// EnumerationPrioritiesAllGet gets info for all priority enumerations
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Enumerations#GET
func (r *Context) EnumerationPrioritiesAllGet() ([]EnumerationPriorityObject, StatusCode, error) {

	var e enumerationPrioritiesAllResult

	status, err := r.Get(
		&e,
		url.URL{
			Path: "/enumerations/issue_priorities.json",
		},
		http.StatusOK,
	)

	return e.Priorities, status, err
}

// EnumerationTimeEntryActivitiesAllGet gets info for all time entry activity enumerations
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Enumerations#GET-2
func (r *Context) EnumerationTimeEntryActivitiesAllGet() ([]EnumerationTimeEntryActivityObject, StatusCode, error) {

	var e enumerationTimeEntryActivitiesAllResult

	status, err := r.Get(
		&e,
		url.URL{
			Path: "/enumerations/time_entry_activities.json",
		},
		http.StatusOK,
	)

	return e.TimeEntryActivities, status, err
}

// EnumerationDocumentCategoriesAllGet gets info for all document category enumerations
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Enumerations#GET-3
func (r *Context) EnumerationDocumentCategoriesAllGet() ([]EnumerationDocumentCategoryObject, StatusCode, error) {

	var e enumerationDocumentCategoriesAllResult

	status, err := r.Get(
		&e,
		url.URL{
			Path: "/enumerations/document_categories.json",
		},
		http.StatusOK,
	)

	return e.DocumentCategories, status, err
}
