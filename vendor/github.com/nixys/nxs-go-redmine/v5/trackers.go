package redmine

import (
	"net/http"
	"net/url"
)

/* Get */

// TrackerObject struct used for trackers get operations
type TrackerObject struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	DefaultStatus         IDName   `json:"default_status"`          // (since 3.0)
	Description           *string  `json:"description"`             // (since 4.2.0)
	EnabledStandardFields []string `json:"enabled_standard_fields"` // (since 5.0.0)
}

/* Internal types */

type trackerAllResult struct {
	Trackers []TrackerObject `json:"trackers"`
}

// TrackerAllGet gets info for all trackers
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Trackers#GET
func (r *Context) TrackerAllGet() ([]TrackerObject, StatusCode, error) {

	var t trackerAllResult

	status, err := r.Get(
		&t,
		url.URL{
			Path: "/trackers.json",
		},
		http.StatusOK,
	)

	return t.Trackers, status, err
}
