package redmine

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// UserStatus defines user status type
type UserStatus int64

// UserNotification defines user notification type
type UserNotification string

type UserInclude string

// UserStatus const
const (
	UserStatusAnonymous  UserStatus = 0
	UserStatusActive     UserStatus = 1
	UserStatusRegistered UserStatus = 2
	UserStatusLocked     UserStatus = 3
)

// UserNotification const
const (
	UserNotificationAll          UserNotification = "all"
	UserNotificationSelected     UserNotification = "selected"
	UserNotificationOnlyMyEvents UserNotification = "only_my_events"
	UserNotificationOnlyAssigned UserNotification = "only_assigned"
	UserNotificationOnlyOwner    UserNotification = "only_owner"
	UserNotificationOnlyNone     UserNotification = "none"
)

const (
	UserIncludeGroups      UserInclude = "groups"
	UserIncludeMemberships UserInclude = "memberships"
)

/* Get */

// UserObject struct used for users get operations
type UserObject struct {
	ID              int64                   `json:"id"`
	Login           string                  `json:"login"`
	Admin           bool                    `json:"admin"`
	FirstName       string                  `json:"firstname"`
	LastName        string                  `json:"lastname"`
	Mail            string                  `json:"mail"`
	CreatedOn       string                  `json:"created_on"`
	LastLoginOn     string                  `json:"last_login_on"`
	PasswdChangedOn string                  `json:"passwd_changed_on"`
	TwofaScheme     *string                 `json:"twofa_scheme"` // has nil value if 2FA not enabled and "totp" string value otherwise
	APIKey          *string                 `json:"api_key"`      // used only: get single user
	Status          *UserStatus             `json:"status"`       // used only: get single user
	CustomFields    []CustomFieldGetObject  `json:"custom_fields"`
	Groups          *[]IDName               `json:"groups"`      // used only: get single user and include specified
	Memberships     *[]UserMembershipObject `json:"memberships"` // used only: get single user and include specified
}

// UserMembershipObject struct used for users get operations
type UserMembershipObject struct {
	ID      int64    `json:"id"`
	Project IDName   `json:"project"`
	Roles   []IDName `json:"roles"`
}

/* Create */

// UserCreate struct used for users create operations
type UserCreate struct {
	User            UserCreateObject `json:"user"`
	SendInformation *bool            `json:"send_information,omitempty"`
}

type UserCreateObject struct {
	Login            string                     `json:"login"`
	FirstName        string                     `json:"firstname"`
	LastName         string                     `json:"lastname"`
	Mail             string                     `json:"mail"`
	Password         *string                    `json:"password,omitempty"`
	AuthSourceID     *int64                     `json:"auth_source_id,omitempty"`
	MailNotification *string                    `json:"mail_notification,omitempty"`
	MustChangePasswd *bool                      `json:"must_change_passwd,omitempty"`
	GeneratePassword *bool                      `json:"generate_password,omitempty"`
	CustomFields     *[]CustomFieldUpdateObject `json:"custom_fields,omitempty"`
}

/* Update */

// UserUpdate struct used for users update operations
type UserUpdate struct {
	User            UserUpdateObject `json:"user"`
	SendInformation *bool            `json:"send_information,omitempty"`
}

type UserUpdateObject struct {
	Login            *string                    `json:"login,omitempty"`
	FirstName        *string                    `json:"firstname,omitempty"`
	LastName         *string                    `json:"lastname,omitempty"`
	Mail             *string                    `json:"mail,omitempty"`
	Password         *string                    `json:"password,omitempty"`
	AuthSourceID     *int64                     `json:"auth_source_id,omitempty"`
	MailNotification *string                    `json:"mail_notification,omitempty"`
	MustChangePasswd *bool                      `json:"must_change_passwd,omitempty"`
	GeneratePassword *bool                      `json:"generate_password,omitempty"`
	CustomFields     *[]CustomFieldUpdateObject `json:"custom_fields,omitempty"`
}

/* Requests */

// UserAllGetRequest contains data for making request to get all users satisfying specified filters
type UserAllGetRequest struct {
	Filters *UserGetRequestFilters
}

// UserMultiGetRequest contains data for making request to get limited users count satisfying specified filters
type UserMultiGetRequest struct {
	Filters *UserGetRequestFilters
	Offset  int64
	Limit   int64
}

// UserSingleGetRequest contains data for making request to get specified user
type UserSingleGetRequest struct {
	Includes []UserInclude
}

// UserCurrentGetRequest contains data for making request to get current user
type UserCurrentGetRequest struct {
	Includes []UserInclude
}

// UserGetRequestFilters contains data for making users get request
type UserGetRequestFilters struct {
	status  *UserStatus
	name    *string
	groupID *int64
}

/* Results */

// UserResult stores users requests processing result
type UserResult struct {
	Users      []UserObject `json:"users"`
	TotalCount int64        `json:"total_count"`
	Offset     int64        `json:"offset"`
	Limit      int64        `json:"limit"`
}

/* Internal types */

type userSingleResult struct {
	User UserObject `json:"user"`
}

func (u UserStatus) String() string {

	status := map[UserStatus]string{
		UserStatusAnonymous:  "anonymous",
		UserStatusActive:     "active",
		UserStatusRegistered: "registered",
		UserStatusLocked:     "locked",
	}

	s, b := status[u]
	if b == false {
		return "unknown"
	}

	return s
}

func (u UserNotification) String() string {
	return string(u)
}

func (ui UserInclude) String() string {
	return string(ui)
}

// UserAllGet gets info for all users satisfying specified filters
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#GET
func (r *Context) UserAllGet(request UserAllGetRequest) (UserResult, StatusCode, error) {

	var (
		users  UserResult
		offset int64
		status StatusCode
	)

	up := request.url()
	up.Set("limit", strconv.FormatInt(limitDefault, 10))

	for {

		var u UserResult

		// m.Offset = offset

		up.Set("offset", strconv.FormatInt(offset, 10))

		s, err := r.Get(
			&u,
			url.URL{
				Path:     "/users.json",
				RawQuery: up.Encode(),
			},
			http.StatusOK,
		)
		if err != nil {
			return users, s, err
		}

		status = s

		users.Users = append(users.Users, u.Users...)

		if offset+u.Limit >= u.TotalCount {
			users.TotalCount = u.TotalCount
			users.Limit = u.TotalCount

			break
		}

		offset += u.Limit
	}

	return users, status, nil
}

// UserMultiGet gets info for multiple users satisfying specified filters
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#GET
func (r *Context) UserMultiGet(request UserMultiGetRequest) (UserResult, StatusCode, error) {

	var u UserResult

	s, err := r.Get(
		&u,
		url.URL{
			Path:     "/users.json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return u, s, err
}

// UserSingleGet gets single user info by specific ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#GET-2
func (r *Context) UserSingleGet(id int64, request UserSingleGetRequest) (UserObject, StatusCode, error) {

	var u userSingleResult

	status, err := r.Get(
		&u,
		url.URL{
			Path:     "/users/" + strconv.FormatInt(id, 10) + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return u.User, status, err
}

// UserCurrentGet gets current user info
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#GET-2
func (r *Context) UserCurrentGet(request UserCurrentGetRequest) (UserObject, StatusCode, error) {

	var u userSingleResult

	status, err := r.Get(
		&u,
		url.URL{
			Path:     "/users/current.json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return u.User, status, err
}

// UserCreate creates new user
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#POST
func (r *Context) UserCreate(user UserCreate) (UserObject, StatusCode, error) {

	var u userSingleResult

	status, err := r.Post(
		user,
		&u,
		url.URL{
			Path: "/users.json",
		},
		http.StatusCreated,
	)

	return u.User, status, err
}

// UserUpdate updates user with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#PUT
func (r *Context) UserUpdate(id int64, user UserUpdate) (StatusCode, error) {

	status, err := r.Put(
		user,
		nil,
		url.URL{
			Path: "/users/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// UserDelete deletes user with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_Users#DELETE
func (r *Context) UserDelete(id int64) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/users/" + strconv.FormatInt(id, 10) + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

func (ur UserAllGetRequest) url() url.Values {

	v := url.Values{}

	if ur.Filters != nil {
		ur.Filters.url(&v)
	}

	return v
}

func (ur UserMultiGetRequest) url() url.Values {

	v := url.Values{}

	if ur.Filters != nil {
		ur.Filters.url(&v)
	}

	v.Set("offset", strconv.FormatInt(ur.Offset, 10))
	v.Set("limit", strconv.FormatInt(ur.Limit, 10))

	return v
}

func (ur UserSingleGetRequest) url() url.Values {

	v := url.Values{}

	if len(ur.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range ur.Includes {
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

func (ur UserCurrentGetRequest) url() url.Values {

	v := url.Values{}

	if len(ur.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range ur.Includes {
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

func UserGetRequestFiltersInit() *UserGetRequestFilters {
	return &UserGetRequestFilters{}
}

func (f *UserGetRequestFilters) StatusSet(s UserStatus) *UserGetRequestFilters {
	f.status = &s
	return f
}

func (f *UserGetRequestFilters) NameSet(n string) *UserGetRequestFilters {
	f.name = &n
	return f
}

func (f *UserGetRequestFilters) GroupIDSet(g int64) *UserGetRequestFilters {
	f.groupID = &g
	return f
}

func (f *UserGetRequestFilters) url(v *url.Values) {

	if f.status != nil {
		v.Set("status", strconv.FormatInt(int64(*f.status), 10))
	}

	if f.name != nil {
		v.Set("name", *f.name)
	}

	if f.groupID != nil {
		v.Set("group_id", strconv.FormatInt(*f.groupID, 10))
	}
}
