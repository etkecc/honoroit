package config

import (
	"time"

	"gitlab.com/etke.cc/go/env"
)

const prefix = "honoroit"

// New config
func New() *Config {
	env.SetPrefix(prefix)
	return &Config{
		Homeserver:     env.String("homeserver", defaultConfig.Homeserver),
		RoomID:         env.String("roomid", defaultConfig.RoomID),
		AllowedUsers:   env.Slice("allowedusers"),
		IgnoredRooms:   env.Slice("ignoredrooms"),
		IgnoreNoThread: env.Bool("ignorenothread"),
		Login:          env.String("login", defaultConfig.Login),
		Password:       env.String("password", defaultConfig.Password),
		LogLevel:       env.String("loglevel", defaultConfig.LogLevel),
		CacheSize:      env.Int("cachesize", defaultConfig.CacheSize),
		NoEncryption:   env.Bool("noencryption"),
		Prefix:         env.String("prefix", defaultConfig.Prefix),
		Port:           env.String("port", defaultConfig.Port),
		DB: DB{
			DSN:     env.String("db.dsn", defaultConfig.DB.DSN),
			Dialect: env.String("db.dialect", defaultConfig.DB.Dialect),
		},
		Monitoring: Monitoring{
			SentryDSN:          env.String("monitoring.sentry.dsn", env.String("sentry", "")),
			SentrySampleRate:   env.Int("monitoring.sentry.rate", env.Int("sentry.rate", 0)),
			HealchecksUUID:     env.String("monitoring.healthchecks.uuid", ""),
			HealthechsDuration: time.Duration(env.Int("monitoring.healthchecks.duration", int(defaultConfig.Monitoring.HealthechsDuration))) * time.Second,
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
