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

[DE] Hallo,
Ihre Nachricht wurde an die Betreiber gesendet.
Bitte bewahren Sie Ruhe und warten Sie auf die Antwort (normalerweise dauert es 1-2 Tage).
Vergessen Sie nicht, dass Instant Messenger derselbe Kommunikationskanal wie E-Mail ist, erwarten Sie also nicht sofort eine Antwort.

[RU] Привет,
Ваше сообщение было отправлено операторам.
Пожалуйста, наберитесь терпения и дождитесь ответа (обычно это занимает 1-2 дня).
Имейте в виду, что мессенджер это такой же канал связи, как и email, так что не стоит ожидать мгновенного ответа.`,
		Join:   "New customer (%s) joined the room",
		Invite: "Customer (%s) invited another user (%s) into the room",
		Leave:  "Customer (%s) left the room",
		Error: `[EN] Something is wrong.
I notified the developers and they are fixing the issue.
Please, try again later or use any other contact method.

[DE] Irgendetwas stimmt nicht.
Ich habe die Entwickler benachrichtigt und sie sind dabei, das Problem zu beheben.
Bitte versuchen Sie es später noch einmal oder verwenden Sie eine andere Kontaktmethode.

[RU] Что-то пошло не так.
Разрабочтики уже в курсе проблемы и заняты ее решением.
Пожалуйста, попробуйте еще раз позже или используйте любой другой способ связи.`,
		EmptyRoom: "The last customer left the room.\nConsider that request closed.",
		Start:     "The customer was invited to the new room. Send messages into that thread and they will be automatically forwarded.",
		Done: `[EN] The operator marked your request as completed.
If you think that it's not done yet, please start another 1:1 chat with me to open a new request.

[DE] Der Betreiber hat Ihre Anfrage als abgeschlossen markiert.
Wenn Sie der Meinung sind, dass es noch nicht fertig ist, starten Sie bitte einen weiteren 1:1-Chat mit mir, um eine neue Anfrage zu öffnen.

[RU] Оператор отметил Ваш запрос как решенный.
Если Вы считаете что он еще не решен, пожалуйста, начните еще один 1:1 чат со мной для открытия нового запроса.`,
		NoEncryption: `[EN] Unfortunately, encryption is disabled to prevent common decryption issues among customers.
Please, start a new un-encrypted chat with me.

[DE] Leider ist die Verschlüsselung deaktiviert, um häufige Entschlüsselungsprobleme bei den Kunden zu vermeiden.
Bitte beginnen Sie einen neuen unverschlüsselten Chat mit mir.

[RU] К сожалению, шифрование отключено чтобы предотвратить частые ошибки с дешифровкой у клиентов.
Пожалуйста, начните новый, незашифрованный чат со мной.`,
	},
}
