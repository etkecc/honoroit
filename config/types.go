package config

// Config of Honoroit
type Config struct {
	// Homeserver url
	Homeserver string
	// Login is a MXID localpart (honoroit - OK, @honoroit:example.com - wrong)
	Login string
	// Password for login/password auth only
	Password string
	// Token for access token auth only (not implemented yet)
	Token string
	// RoomID where threads will be created
	RoomID string
	// LogLevel for logger
	LogLevel string

	Text Text
}

// Text messages
type Text struct {
	// Greetings message sent to customer on first contact
	Greetings string
	// Error message sent to customer if something goes wrong
	Error string
	// EmptyRoom message sent to backoffice/threads room when customer left his room
	EmptyRoom string
	// Done message sent to customer when request marked as done in the threads room
	Done string
}
