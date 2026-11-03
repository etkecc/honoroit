package redmine

import (
	"io"
	"net/url"
	"sync"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

// API is an interface for Redmine API
type API interface {
	ProjectSingleGet(identifier string, req redmine.ProjectSingleGetRequest) (redmine.ProjectObject, redmine.StatusCode, error)
	UserCurrentGet(req redmine.UserCurrentGetRequest) (redmine.UserObject, redmine.StatusCode, error)
	IssueCreate(req redmine.IssueCreate) (redmine.IssueObject, redmine.StatusCode, error)
	IssueUpdate(id int64, req redmine.IssueUpdate) (redmine.StatusCode, error)
	IssueSingleGet(id int64, req redmine.IssueSingleGetRequest) (redmine.IssueObject, redmine.StatusCode, error)
	IssueDelete(id int64) (redmine.StatusCode, error)
	AttachmentUpload(filePath string) (redmine.AttachmentUploadObject, redmine.StatusCode, error)
	AttachmentUploadStream(f io.Reader, fileName string) (redmine.AttachmentUploadObject, redmine.StatusCode, error)
	Del(in, out any, uri url.URL, statusExpected redmine.StatusCode) (redmine.StatusCode, error)
	Post(in, out any, uri url.URL, statusExpected redmine.StatusCode) (redmine.StatusCode, error)
}

// Redmine is a Redmine client
type Redmine struct {
	wg  sync.WaitGroup
	cfg *Config
}

// New creates a new Redmine client
func New(options ...Option) (*Redmine, error) {
	cfg := NewConfig(options...)
	r := &Redmine{cfg: cfg}
	if !cfg.Enabled() {
		return r, nil
	}

	if cfg.ProjectID == 0 {
		if err := r.UpdateProject(); err != nil {
			return r, err
		}
	}

	if cfg.UserID == 0 {
		if err := r.UpdateUser(); err != nil {
			return r, err
		}
	}
	return r, nil
}

// GetAPI returns the underlying nxs-go-redmine object for unexposed methods; can return nil on a type-cast mismatch.
func (r *Redmine) GetAPI() *redmine.Context {
	if r.cfg == nil {
		return nil
	}
	if r.cfg.api == nil {
		return nil
	}

	typed, ok := r.cfg.api.(*redmine.Context)
	if !ok {
		return nil
	}
	return typed
}

// GetHost returns the Redmine host
func (r *Redmine) GetHost() string {
	return r.cfg.Host
}

// GetAPIKey returns the Redmine API key
func (r *Redmine) GetAPIKey() string {
	return r.cfg.APIKey
}

// GetProjectIdentifier returns the Redmine project identifier
func (r *Redmine) GetProjectIdentifier() string {
	return r.cfg.ProjectIdentifier
}

// GetProjectID returns the Redmine project ID
func (r *Redmine) GetProjectID() int64 {
	return r.cfg.ProjectID
}

// GetUserID returns the Redmine user ID
func (r *Redmine) GetUserID() int64 {
	return r.cfg.UserID
}

// GetTrackerID returns the Redmine tracker ID
func (r *Redmine) GetTrackerID() int64 {
	return r.cfg.TrackerID
}

// GetNewStatusID returns the Redmine new status ID
func (r *Redmine) GetWaitingForOperatorStatusID() int64 {
	return r.cfg.WaitingForOperatorStatusID
}

// GetWaitingForCustomerStatusID returns the Redmine waiting for customer status ID
func (r *Redmine) GetWaitingForCustomerStatusID() int64 {
	return r.cfg.WaitingForCustomerStatusID
}

// GetDoneStatusID returns the Redmine done status ID
func (r *Redmine) GetDoneStatusID() int64 {
	return r.cfg.DoneStatusID
}

// Enabled returns true if the Redmine client is enabled
func (r *Redmine) Enabled() bool {
	return r.cfg.Enabled()
}

// Configure applies config options at runtime; call UpdateUser()/UpdateProject() after an API key/host/project change.
func (r *Redmine) Configure(options ...Option) *Redmine {
	r.cfg.apply(options...)
	return r
}

// UpdateUser updates the current user ID; call after changing the API key and/or host.
func (r *Redmine) UpdateUser() error {
	user, err := RetryResult(r.cfg.Log, func() (redmine.UserObject, redmine.StatusCode, error) {
		return r.cfg.api.UserCurrentGet(redmine.UserCurrentGetRequest{})
	})
	if err != nil {
		return err
	}
	r.cfg.UserID = user.ID
	return nil
}

// UpdateProject updates the project ID; call after changing the project identifier.
func (r *Redmine) UpdateProject() error {
	project, err := RetryResult(r.cfg.Log, func() (redmine.ProjectObject, redmine.StatusCode, error) {
		return r.cfg.api.ProjectSingleGet(r.cfg.ProjectIdentifier, redmine.ProjectSingleGetRequest{})
	})
	if err != nil {
		return err
	}
	r.cfg.ProjectID = project.ID
	return nil
}

// Shutdown waits for all goroutines to finish
func (r *Redmine) Shutdown() {
	r.wg.Wait()
}
