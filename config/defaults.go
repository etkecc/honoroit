package config

var defaultConfig = &Config{
	Prefix:    "!ho",
	LogLevel:  "INFO",
	CacheSize: 2000,
	Port:      "8080",
	DB: DB{
		DSN:     "/tmp/honoroit.db",
		Dialect: "sqlite3",
	},
	Monitoring: Monitoring{
		SentrySampleRate:   20,
		HealthechsDuration: 5,
	},
	Text: Text{
		PrefixOpen: "[OPEN]",
		PrefixDone: "[DONE]",
		Greetings: `Hello,
Your message was sent to operators.
Please, keep calm and wait for the answer (usually, it takes 1-2 days).
Don't forget that instant messenger is the same communication channel as email, so don't expect an instant response.
Please be advised that requests that are older than 7 days are eligible for automatic removal on our end.`,
		Join:   "New customer (%s) joined the room",
		Invite: "Customer (%s) invited another user (%s) into the room",
		Leave:  "Customer (%s) left the room",
		Error: `Something is wrong.
I notified the developers and they are fixing the issue.
Please, try again later or use any other contact method.`,
		EmptyRoom: "The last customer left the room.\nConsider that request closed.",
		Start:     "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
		Count:     "Request has been counted.",
		Done: `The operator marked your request as completed.
If you think that it's not done yet, please start another 1:1 chat with me to open a new request.`,
		NoEncryption: `Unfortunately, encryption is disabled to prevent common decryption issues among customers.
Please, start a new un-encrypted chat with me.`,
	},
}
