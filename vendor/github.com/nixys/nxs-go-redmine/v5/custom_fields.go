package redmine

import (
	"net/http"
	"net/url"
)

/* Get */

// CustomFieldObject struct used for custom fields get operations
type CustomFieldObject struct {
	ID             int64                             `json:"id"`
	Name           string                            `json:"name"`
	CustomizedType string                            `json:"customized_type"`
	FieldFormat    string                            `json:"field_format"`
	Regexp         string                            `json:"regexp"`
	MinLength      int64                             `json:"min_length"`
	MaxLength      int64                             `json:"max_length"`
	IsRequired     bool                              `json:"is_required"`
	IsFilter       bool                              `json:"is_filter"`
	Searchable     bool                              `json:"searchable"`
	Multiple       bool                              `json:"multiple"`
	DefaultValue   *string                           `json:"default_value"`
	Visible        bool                              `json:"visible"`
	Trackers       []IDName                          `json:"trackers"`
	PossibleValues *[]CustomFieldPossibleValueObject `json:"possible_values"`
	Roles          []IDName                          `json:"roles"`
}

// CustomFieldPossibleValueObject struct used for custom fields get operations
type CustomFieldPossibleValueObject struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// CustomFieldGetObject struct used for custom fields get operations in other methods
type CustomFieldGetObject struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Multiple *bool     `json:"multiple"`
	Value    *[]string `json:"value"`
}

/* Update */

// CustomFieldUpdateObject struct used for custom fields insert and update operations in other methods
type CustomFieldUpdateObject struct {
	ID    int64       `json:"id"`
	Value interface{} `json:"value"` // can be a string or strings slice
}

/* Internal types */

type customFieldAllResult struct {
	CustomFields []CustomFieldObject `json:"custom_fields"`
}

// CustomFieldAllGet gets info for all custom fields
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_CustomFields#GET
func (r *Context) CustomFieldAllGet() ([]CustomFieldObject, StatusCode, error) {

	var c customFieldAllResult

	status, err := r.Get(
		&c,
		url.URL{
			Path: "/custom_fields.json",
		},
		http.StatusOK,
	)

	return c.CustomFields, status, err
}
