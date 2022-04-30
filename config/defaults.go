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
		Greetings: `[EN] Hello,
Your message was sent to operators.
Please, keep calm and wait for the answer (usually, it takes 1-2 days).
Don't forget that instant messenger is the same communication channel as email, so don't expect an instant response.

[RU] Привет,
Твое сообщение было отправлено операторам.
Пожалуйста, наберись терпения и дождись ответа (обычно это занимает 1-2 дня).
Имей в виду, что мессенджер это такой же канал связи, как и email, так что не стоит ожидать мгновенного ответа.`,
		Join:   "New customer (%s) joined the room",
		Invite: "Customer (%s) invited another user (%s) into the room",
		Leave:  "Customer (%s) left the room",
		Error: `[EN] Something is wrong.
I notified the developers and they are fixing the issue.
Please, try again later or use any other contact method.

[RU] Что-то пошло не так.
Разрабочтики уже в курсе проблемы и заняты ее решением.
Пожалуйста, попробуй еще раз позже или используй любой другой способ связи.`,
		EmptyRoom: "The last customer left the room.\nConsider that request closed.",
		Start:     "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
		Done: `[EN] The operator marked your request as completed.
If you think that it's not done yet, please start another 1:1 chat with me to open a new request.

[RU] Оператор отметил твой запрос как решенный.
Если ты считаешь что он еще не решен, пожалуйста, начни еще один 1:1 чат со мной для открытия нового запроса.`,
	},
}
