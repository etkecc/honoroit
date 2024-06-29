package redmine

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type GroupInclude string

const (
	GroupIncludeUsers       GroupInclude = "users"       // used only: get single user
	GroupIncludeMemberships GroupInclude = "memberships" // used only: get single user
)

/* Get */

// GroupObject struct used for groups get operations
type GroupObject struct {
	ID          int64                    `json:"id"`
	Name        string                   `json:"name"`
	Users       *[]IDName                `json:"users"`       // used only: get single user and include specified
	Memberships *[]GroupMembershipObject `json:"memberships"` // used only: get single user and include specified
}

// GroupMembershipObject struct used for groups get operations
type GroupMembershipObject struct {
	ID      int64    `json:"id"`
	Project IDName   `json:"project"`
	Roles   []IDName `json:"roles"`
}

/* Create */

// GroupCreate struct used for groups create operations
type GroupCreate struct {
	Group GroupCreateObject `json:"group"`
}

type GroupCreateObject struct {
	Name    string   `json:"name"`
	UserIDs *[]int64 `json:"user_ids,omitempty"`
}

/* Update */

// GroupUpdate struct used for groups update operations
type GroupUpdate struct {
	Group GroupUpdateObject `json:"group"`
}

type GroupUpdateObject struct {
	Name    *string  `json:"name,omitempty"`
	UserIDs *[]int64 `json:"user_ids,omitempty"`
}

/* Add user */

// GroupAddUserObject struct used for add new user into group
type GroupAddUserObject struct {
	UserID int64 `json:"user_id"`
}

/* Requests */

// GroupMultiGetRequest contains data for making request to get limited groups count
type GroupMultiGetRequest struct {
	Offset int64
	Limit  int64
}

// GroupSingleGetRequest contains data for making request to get specified group
type GroupSingleGetRequest struct {
	Includes []GroupInclude
}

/* Results */

// GroupResult stores groups requests processing result
type GroupResult struct {
	Groups     []GroupObject `json:"groups"`
	TotalCount int64         `json:"total_count"`
	Offset     int64         `json:"offset"`
	Limit      int64         `json:"limit"`
}

/* Internal types */

type groupSingleResult struct {
	Group GroupObject `json:"group"`
}

func (gi GroupInclude) String() string {
	return string(gi)
}

// GroupAllGet gets info for all groups
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#GET
func (r *Context) GroupAllGet() (GroupResult, StatusCode, error) {

	var (
		groups GroupResult
		offset int64
		status StatusCode
	)

	for {

		g, s, err := r.GroupMultiGet(
			GroupMultiGetRequest{
				Limit:  limitDefault,
				Offset: offset,
			},
		)
		if err != nil {
			return groups, s, err
		}

		status = s

		groups.Groups = append(groups.Groups, g.Groups...)

		if offset+g.Limit >= g.TotalCount {
			groups.TotalCount = g.TotalCount
			groups.Limit = g.TotalCount

			break
		}

		offset += g.Limit
	}

	return groups, status, nil
}

// GroupMultiGet gets info for multiple groups
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#GET
func (r *Context) GroupMultiGet(request GroupMultiGetRequest) (GroupResult, StatusCode, error) {

	var g GroupResult

	s, err := r.Get(
		&g,
		url.URL{
			Path:     "/groups.json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return g, s, err
}

// GroupSingleGet gets single group info by specific ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#GET-2
func (r *Context) GroupSingleGet(id int64, request GroupSingleGetRequest) (GroupObject, StatusCode, error) {

	var g groupSingleResult

	status, err := r.Get(
		&g,
		url.URL{
			Path:     "/groups/" + strconv.FormatInt(id, 10) + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return g.Group, status, err
}

// GroupCreate creates new group
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#POST
func (r *Context) GroupCreate(group GroupCreate) (GroupObject, StatusCode, error) {

	var g groupSingleResult

	status, err := r.Post(
		group,
		&g,
		url.URL{
			Path: "/groups.json",
		},
		http.StatusCreated,
	)

	return g.Group, status, err
}

// GroupUpdate updates group with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#PUT
func (r *Context) GroupUpdate(id int64, group GroupUpdate) (StatusCode, error) {

	status, err := r.Put(
		group,
		nil,
		url.URL{
			Path: "/groups/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// GroupDelete deletes group with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#DELETE
func (r *Context) GroupDelete(id int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/groups/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// GroupAddUser adds new user into group with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#POST-2
func (r *Context) GroupAddUser(id int64, group GroupAddUserObject) (StatusCode, error) {

	status, err := r.Post(
		group,
		nil,
		url.URL{
			Path: "/groups/" + strconv.FormatInt(id, 10) + "/users.json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// GroupDeleteUser deletes user from group with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Groups#DELETE-2
func (r *Context) GroupDeleteUser(id int64, userID int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/groups/" + strconv.FormatInt(id, 10) + "/users/" + strconv.FormatInt(userID, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

func (gr GroupMultiGetRequest) url() url.Values {

	v := url.Values{}

	v.Set("offset", strconv.FormatInt(gr.Offset, 10))
	v.Set("limit", strconv.FormatInt(gr.Limit, 10))

	return v
}

func (gr GroupSingleGetRequest) url() url.Values {

	v := url.Values{}

	if len(gr.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range gr.Includes {
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
