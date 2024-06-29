package config

import (
	"time"

	echobasicauth "gitlab.com/etke.cc/go/echo-basic-auth"
	"gitlab.com/etke.cc/go/env"
)

const prefix = "honoroit"

// New config
func New() *Config {
	env.SetPrefix(prefix)
	return &Config{
		Homeserver:   env.String("homeserver", defaultConfig.Homeserver),
		RoomID:       env.String("roomid", defaultConfig.RoomID),
		Login:        env.String("login", defaultConfig.Login),
		Password:     env.String("password", defaultConfig.Password),
		SharedSecret: env.String("sharedsecret", defaultConfig.SharedSecret),
		DataSecret:   env.String("data.secret", defaultConfig.DataSecret),
		LogLevel:     env.String("loglevel", defaultConfig.LogLevel),
		CacheSize:    env.Int("cachesize", defaultConfig.CacheSize),
		Prefix:       env.String("prefix", defaultConfig.Prefix),
		Port:         env.String("port", defaultConfig.Port),
		DB: DB{
			DSN:     env.String("db.dsn", defaultConfig.DB.DSN),
			Dialect: env.String("db.dialect", defaultConfig.DB.Dialect),
		},
		Auth: Auth{
			Metrics: &echobasicauth.Auth{
				Login:    env.String("auth.metrics.login"),
				Password: env.String("auth.metrics.password"),
				IPs:      env.Slice("auth.metrics.ips"),
			},
			PSD: PSDAuth{
				URL:      env.String("auth.psd.url"),
				Login:    env.String("auth.psd.login"),
				Password: env.String("auth.psd.password"),
			},
		},
		Redmine: Redmine{
			Host:             env.String("redmine.host"),
			APIKey:           env.String("redmine.apikey"),
			ProjectID:        env.String("redmine.project"),
			TrackerID:        env.Int("redmine.trackerid"),
			NewStatus:        env.Int("redmine.newstatusid"),
			InProgressStatus: env.Int("redmine.inprogressstatusid"),
			DoneStatus:       env.Int("redmine.donestatusid"),
		},
		Monitoring: Monitoring{
			SentryDSN:            env.String("monitoring.sentry.dsn", env.String("sentry", "")),
			SentrySampleRate:     env.Int("monitoring.sentry.rate", env.Int("sentry.rate", 0)),
			HealthchecksURL:      env.String("monitoring.healthchecks.url", defaultConfig.Monitoring.HealthchecksURL),
			HealthchecksUUID:     env.String("monitoring.healthchecks.uuid", ""),
			HealthchecksDuration: time.Duration(env.Int("monitoring.healthchecks.duration", int(defaultConfig.Monitoring.HealthchecksDuration))) * time.Second,
		},
	}
}
