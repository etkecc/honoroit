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

// GetAPI returns the underlying API object from the github.com/nixys/nxs-go-redmine package (actual version may vary)
// it's useful for calling methods that are not exposed by this package
// WARNING: it CAN return nil (API interface is defined for tests only, but this method does type casting, so if the type is wrong, it will return nil)
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

// Configure applies the new configuration options in runtime
// It is advisable to call UpdateUser() (if API key and/or host was changed), and UpdatePorject() (if project identifier was changed) after this method
func (r *Redmine) Configure(options ...Option) *Redmine {
	r.cfg.apply(options...)
	return r
}

// UpdateUser updates the current user ID,
// it should be called after changing the API key and/or host
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

// UpdateProject updates the project ID,
// it should be called after changing the project identifier
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
