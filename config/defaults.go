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
}
