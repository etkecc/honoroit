package config

var defaultConfig = &Config{
	LogLevel: "INFO",
	TTL:      1,
	DB: DB{
		DSN:     "/tmp/honoroit.db",
		Dialect: "sqlite3",
	},
	Text: Text{
		Greetings: "Hello\nyour message was sent to developers. Please, keep calm and wait for answer, usually it takes 1-2 days.",
		Error:     "Something is wrong.\nI already notified developers, they're fixing the issue.\n\nPlease, try again later or use any other contact method.",
		EmptyRoom: "Customer left the room.\nConsider that request as closed.",
		Done:      "Developer marked your request as completed.\nIf you think that it's not done yet, please, start another 1:1 chat with me to open a new request.",
	},
}
