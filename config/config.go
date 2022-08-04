package config

import (
	"gitlab.com/etke.cc/go/env"
)

const prefix = "honoroit"

// New config
func New() *Config {
	env.SetPrefix(prefix)
	return &Config{
		Homeserver:     env.String("homeserver", defaultConfig.Homeserver),
		RoomID:         env.String("roomid", defaultConfig.RoomID),
		IgnoredRooms:   env.Slice("ignoredrooms"),
		IgnoreNoThread: env.Bool("ignorenothread"),
		Login:          env.String("login", defaultConfig.Login),
		Password:       env.String("password", defaultConfig.Password),
		Sentry:         env.String("sentry", defaultConfig.Sentry),
		LogLevel:       env.String("loglevel", defaultConfig.LogLevel),
		CacheSize:      env.Int("cachesize", defaultConfig.CacheSize),
		NoEncryption:   env.Bool("noencryption"),
		Prefix:         env.String("prefix", defaultConfig.Prefix),
		DB: DB{
			DSN:     env.String("db.dsn", defaultConfig.DB.DSN),
			Dialect: env.String("db.dialect", defaultConfig.DB.Dialect),
		},
		Text: Text{
			PrefixOpen:   env.String("text.prefix.open", defaultConfig.Text.PrefixOpen),
			PrefixDone:   env.String("text.prefix.done", defaultConfig.Text.PrefixDone),
			NoEncryption: env.String("text.noencryption", defaultConfig.Text.NoEncryption),
			Greetings:    env.String("text.greetings", defaultConfig.Text.Greetings),
			Join:         env.String("text.join", defaultConfig.Text.Join),
			Invite:       env.String("text.invite", defaultConfig.Text.Invite),
			Leave:        env.String("text.leave", defaultConfig.Text.Leave),
			Error:        env.String("text.error", defaultConfig.Text.Error),
			EmptyRoom:    env.String("text.emptyroom", defaultConfig.Text.EmptyRoom),
			Start:        env.String("text.start", defaultConfig.Text.Start),
			Done:         env.String("text.done", defaultConfig.Text.Done),
		},
	}
}
