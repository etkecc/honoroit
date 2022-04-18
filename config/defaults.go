package config

var defaultConfig = &Config{
	Prefix:    "!ho",
	LogLevel:  "INFO",
	CacheSize: 2000,
	DB: DB{
		DSN:     "/tmp/honoroit.db",
		Dialect: "sqlite3",
	},
	Text: Text{
		PrefixOpen: "[OPEN]",
		PrefixDone: "[DONE]",
		Greetings:  "Hello,\nYour message was sent to operators.\nPlease, keep calm and wait for the answer (usually, it takes 1-2 days).\nDon't forget that instant messenger is the same communication channel as email, so don't expect an instant response.",
		Join:       "New customer (%s) joined the room",
		Invite:     "Customer (%s) invited another user (%s) into the room",
		Leave:      "Customer (%s) left the room",
		Error:      "Something is wrong.\nI notified the developers and they are fixing the issue.\n\nPlease, try again later or use any other contact method.",
		EmptyRoom:  "The last customer left the room.\nConsider that request closed.",
		Start:      "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
		Done:       "The operator marked your request as completed.\nIf you think that it's not done yet, please start another 1:1 chat with me to open a new request.",
	},
}
