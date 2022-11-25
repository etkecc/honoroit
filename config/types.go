package config

import "time"

// Config of Honoroit
type Config struct {
	// Homeserver url
	Homeserver string
	// Login is a MXID localpart (honoroit - OK, @honoroit:example.com - wrong)
	Login string
	// Password for login/password auth only
	Password string
	// RoomID where threads will be created
	RoomID string
	// AllowedUsers is list of wildcard rules to allow requests only from specific users
	AllowedUsers []string
	// IgnoredRooms list of room IDs to ignore
	IgnoredRooms []string
	// IgnoreNoThread mode completely ignores any messages sent outside of thread
	IgnoreNoThread bool
	// Prefix for honoroit commands
	Prefix string
	// LogLevel for logger
	LogLevel string
	// CacheSize max amount of items in cache
	CacheSize int

	// NoEncryption disabled matrix e2e encryption support
	NoEncryption bool

	// Text messages
	Text Text

	// DB config
	DB DB

	// Monitoring config
	Monitoring Monitoring
}

// DB config
type DB struct {
	// DSN is a database connection string
	DSN string
	// Dialect of the db, allowed values: postgres, sqlite3
	Dialect string
}

// Monitoring config
type Monitoring struct {
	SentryDSN          string
	SentrySampleRate   int
	HealchecksUUID     string
	HealthechsDuration time.Duration
}

// Text messages
type Text struct {
	// PrefixOpen is a prefix added to new thread topics
	PrefixOpen string
	// PrefixDone is a prefix added to threads marked as done/closed
	PrefixDone string

	// NoEncryption message sent to customer when encryption disabled and customer tries to use encrypted chat
	NoEncryption string
	// Greetings message sent to customer on first contact
	Greetings string
	// Join message sent to backoffice/threads room when customer joins a room
	Join string
	// Invite message sent to backoffice/threads room when customer invites somebody into a room
	Invite string
	// Leave message sent to backoffice/threads room when a customer leaves a room
	Leave string
	// Error message sent to customer if something goes wrong
	Error string
	// EmptyRoom message sent to backoffice/threads room when the last customer left a room
	EmptyRoom string
	// Start message that sent into the read as result of the "start" command
	Start string
	// Done message sent to customer when request marked as done in the threads room
	Done string
}
