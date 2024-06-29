package redmine

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type WikiInclude string

const (
	WikiIncludeAttachments WikiInclude = "attachments"
)

/* Get */

// WikiMultiObject struct used for wikies all get operations
type WikiMultiObject struct {
	Title     string            `json:"title"`
	Parent    *WikiParentObject `json:"parent"`
	Version   int64             `json:"version"`
	CreatedOn string            `json:"created_on"`
	UpdatedOn string            `json:"updated_on"`
}

// WikiObject struct used for wiki get operations
type WikiObject struct {
	Title       string              `json:"title"`
	Parent      *WikiParentObject   `json:"parent"`
	Text        string              `json:"text"`
	Version     int64               `json:"version"`
	Author      IDName              `json:"author"`
	Comments    string              `json:"comments"`
	CreatedOn   string              `json:"created_on"`
	UpdatedOn   string              `json:"updated_on"`
	Attachments *[]AttachmentObject `json:"attachments"`
}

// WikiParentObject struct used for wikies get operations
type WikiParentObject struct {
	Title string `json:"title"`
}

/* Create */

// WikiCreate struct used for wiki create operations
type WikiCreate struct {
	WikiPage WikiCreateObject `json:"wiki_page"`
}

type WikiCreateObject struct {
	Text     string                    `json:"text"`
	Comments *string                   `json:"comments,omitempty"`
	Uploads  *[]AttachmentUploadObject `json:"uploads,omitempty"`
}

/* Update */

// WikiUpdate struct used for wiki update operations
type WikiUpdate struct {
	WikiPage WikiUpdateObject `json:"wiki_page"`
}

type WikiUpdateObject struct {
	Text     string                    `json:"text"`
	Comments *string                   `json:"comments,omitempty"`
	Version  *int64                    `json:"version,omitempty"`
	Uploads  *[]AttachmentUploadObject `json:"uploads,omitempty"`
}

/* Requests */

// WikiSingleGetRequest contains data for making request to get specified wiki
type WikiSingleGetRequest struct {
	Includes []WikiInclude
}

/* Internal types */

type wikiAllResult struct {
	WikiPages []WikiMultiObject `json:"wiki_pages"`
}

type wikiSingleResult struct {
	WikiPage WikiObject `json:"wiki_page"`
}

func (wi WikiInclude) String() string {
	return string(wi)
}

// WikiAllGet gets info for all wikies for project with specified ID
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Getting-the-pages-list-of-a-wiki
func (r *Context) WikiAllGet(projectID string) ([]WikiMultiObject, StatusCode, error) {

	var w wikiAllResult

	status, err := r.Get(
		&w,
		url.URL{
			Path: "/projects/" + projectID + "/wiki/index.json",
		},
		http.StatusOK,
	)

	return w.WikiPages, status, err
}

// WikiSingleGet gets single wiki info by specific project ID and wiki title
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Getting-a-wiki-page
func (r *Context) WikiSingleGet(projectID, wikiTitle string, request WikiSingleGetRequest) (WikiObject, StatusCode, error) {

	var w wikiSingleResult

	status, err := r.Get(
		&w,
		url.URL{
			Path:     "/projects/" + projectID + "/wiki/" + wikiTitle + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return w.WikiPage, status, err
}

// WikiSingleVersionGet gets single wiki info by specific project ID, wiki title and version
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Getting-an-old-version-of-a-wiki-page
func (r *Context) WikiSingleVersionGet(projectID, wikiTitle string, version int64, request WikiSingleGetRequest) (WikiObject, StatusCode, error) {

	var w wikiSingleResult

	status, err := r.Get(
		&w,
		url.URL{
			Path:     "/projects/" + projectID + "/wiki/" + wikiTitle + "/" + strconv.FormatInt(version, 10) + ".json",
			RawQuery: request.url().Encode(),
		},
		http.StatusOK,
	)

	return w.WikiPage, status, err
}

// WikiCreate creates new wiki
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Creating-or-updating-a-wiki-page
func (r *Context) WikiCreate(projectID, wikiTitle string, wiki WikiCreate) (WikiObject, StatusCode, error) {

	var w wikiSingleResult

	status, err := r.Put(
		wiki,
		&w,
		url.URL{
			Path: "/projects/" + projectID + "/wiki/" + wikiTitle + ".json",
		},
		http.StatusCreated,
	)

	return w.WikiPage, status, err
}

// WikiUpdate updates wiki page
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Creating-or-updating-a-wiki-page
func (r *Context) WikiUpdate(projectID, wikiTitle string, wiki WikiUpdate) (StatusCode, error) {

	status, err := r.Put(
		wiki,
		nil,
		url.URL{
			Path: "/projects/" + projectID + "/wiki/" + wikiTitle + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

// WikiDelete deletes wiki with specified project ID and title
//
// see: https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages#Deleting-a-wiki-page
func (r *Context) WikiDelete(projectID, wikiTitle string) (StatusCode, error) {

	status, err := r.Del(
		nil,
		nil,
		url.URL{
			Path: "/projects/" + projectID + "/wiki/" + wikiTitle + ".json",
		},
		http.StatusNoContent,
	)

	return status, err
}

func (wr WikiSingleGetRequest) url() url.Values {

	v := url.Values{}

	if len(wr.Includes) > 0 {
		v.Set(
			"include",
			strings.Join(
				func() []string {
					var is []string
					for _, i := range wr.Includes {
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
