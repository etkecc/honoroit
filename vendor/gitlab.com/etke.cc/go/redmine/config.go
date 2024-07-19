package redmine

import (
	redmine "github.com/nixys/nxs-go-redmine/v5"
	"github.com/rs/zerolog"
)

// Config is a configuration for Redmine
type Config struct {
	api                        API             // API interface
	Log                        *zerolog.Logger // Logger, defaults to discard
	Host                       string          // Redmine host, e.g. "https://redmine.example.com"
	APIKey                     string          // Redmine REST API Key
	ProjectIdentifier          string          // Redmine project identifier, e.g. "my-project"
	ProjectID                  int64           // Redmine project ID, e.g. 123. This is set automatically if you use ProjectIdentifier
	UserID                     int64           // Current redmine user ID, e.g. 123. This is set automatically
	TrackerID                  int64           // Task Tracker ID, e.g. 1
	WaitingForOperatorStatusID int64           // Status ID for "Waiting for operator", e.g. 1
	WaitingForCustomerStatusID int64           // Status ID for "Waiting for customer", e.g. 2
	DoneStatusID               int64           // Status ID for "Done", e.g. 3
}

// Option is a functional option for Config
type Option func(*Config)

// NewConfig creates a new Config
func NewConfig(opts ...Option) *Config {
	return (&Config{}).apply(opts...)
}

func (cfg *Config) apply(opts ...Option) *Config {
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Log == nil {
		discardLogger := zerolog.New(nil)
		cfg.Log = &discardLogger
	}

	if cfg.api == nil {
		cfg.api = redmine.Init(redmine.Settings{
			Endpoint: cfg.Host,
			APIKey:   cfg.APIKey,
		})
	}

	return cfg
}

func (cfg *Config) Enabled() bool {
	return cfg.Host != "" &&
		cfg.APIKey != "" &&
		(cfg.ProjectIdentifier != "" || cfg.ProjectID != 0) &&
		cfg.TrackerID != 0 &&
		cfg.WaitingForOperatorStatusID != 0 &&
		cfg.WaitingForCustomerStatusID != 0 &&
		cfg.DoneStatusID != 0
}

// withAPI sets the API interface
func withAPI(api API) Option {
	return func(c *Config) {
		c.api = api
	}
}

// WithLog sets the logger
func WithLog(log *zerolog.Logger) Option {
	return func(c *Config) {
		c.Log = log
	}
}

// WithHost sets the host
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithAPIKey sets the API key
func WithAPIKey(apikey string) Option {
	return func(c *Config) {
		c.APIKey = apikey
	}
}

// WithUserID sets the user ID
func WithUserID(userID int) Option {
	return func(c *Config) {
		c.UserID = int64(userID)
	}
}

// WithProjectIdentifier sets the project identifier
func WithProjectIdentifier(projectIdentifier string) Option {
	return func(c *Config) {
		c.ProjectIdentifier = projectIdentifier
	}
}

// WithProjectID sets the project ID. This is set automatically if you use ProjectIdentifier
func WithProjectID(projectID int) Option {
	return func(c *Config) {
		c.ProjectID = int64(projectID)
	}
}

// WithTrackerID sets the tracker ID
func WithTrackerID(trackerID int) Option {
	return func(c *Config) {
		c.TrackerID = int64(trackerID)
	}
}

// WithWaitingForOperatorStatusID sets the waiting for operator status ID
func WithWaitingForOperatorStatusID(waitingForOperatorStatusID int) Option {
	return func(c *Config) {
		c.WaitingForOperatorStatusID = int64(waitingForOperatorStatusID)
	}
}

// WithWaitingForCustomerStatusID sets the waiting for customer status ID
func WithWaitingForCustomerStatusID(waitingForCustomerStatusID int) Option {
	return func(c *Config) {
		c.WaitingForCustomerStatusID = int64(waitingForCustomerStatusID)
	}
}

// WithDoneStatusID sets the done status ID
func WithDoneStatusID(doneStatusID int) Option {
	return func(c *Config) {
		c.DoneStatusID = int64(doneStatusID)
	}
}
