package config

import (
	"os"
	"strings"
)

const prefix = "honoroit"

func env(shortkey string, defaultValue string) string {
	key := strings.ToUpper(prefix + "_" + strings.ReplaceAll(shortkey, ".", "_"))
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}

	return value
}

// New config
func New() *Config {
	return &Config{
		Homeserver: env("homeserver", defaultConfig.Homeserver),
		RoomID:     env("roomid", defaultConfig.RoomID),
		Login:      env("login", defaultConfig.Login),
		Password:   env("password", defaultConfig.Password),
		Sentry:     env("sentry", defaultConfig.Sentry),
		LogLevel:   env("loglevel", defaultConfig.LogLevel),
		Prefix:     env("prefix", defaultConfig.Prefix),
		DB: DB{
			DSN:     env("db.dsn", defaultConfig.DB.DSN),
			Dialect: env("db.dialect", defaultConfig.DB.Dialect),
		},
		Mail: Mail{
			IMAPhost: env("mail.imap.host", defaultConfig.Mail.IMAPhost),
			IMAPport: env("mail.imap.port", defaultConfig.Mail.IMAPport),
			SMTPhost: env("mail.smtp.host", defaultConfig.Mail.SMTPhost),
			SMTPport: env("mail.smtp.port", defaultConfig.Mail.SMTPport),
			Login:    env("mail.login", defaultConfig.Mail.Login),
			Password: env("mail.password", defaultConfig.Mail.Password),
			Email:    env("mail.email", defaultConfig.Mail.Email),
			Mailbox:  env("mail.mailbox", defaultConfig.Mail.Mailbox),
			Sentbox:  env("mail.sentbox", defaultConfig.Mail.Sentbox),
		},
		Text: Text{
			Greetings: env("text.greetings", defaultConfig.Text.Greetings),
			Error:     env("text.error", defaultConfig.Text.Error),
			EmptyRoom: env("text.emptyroom", defaultConfig.Text.EmptyRoom),
			Done:      env("text.done", defaultConfig.Text.Done),
		},
	}
}
