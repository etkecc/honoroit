package config

import (
	"os"
	"strings"
)

const prefix = "honoroit"

func env(shortkey string) string {
	key := strings.ToUpper(prefix + "_" + strings.ReplaceAll(shortkey, ".", "_"))
	return strings.TrimSpace(os.Getenv(key))
}

// New config
func New() *Config {
	config := defaultConfig

	// matrix
	config.Homeserver = env("homeserver")
	config.RoomID = env("roomid")

	// matrix::auth
	config.Login = env("login")
	config.Password = env("password")
	config.Token = env("token")

	if level := env("loglevel"); level != "" {
		config.LogLevel = level
	}

	// text
	if txt := env("text.greetings"); txt != "" {
		config.Text.Greetings = txt
	}

	if txt := env("text.error"); txt != "" {
		config.Text.Error = txt
	}

	if txt := env("text.emptyroom"); txt != "" {
		config.Text.EmptyRoom = txt
	}

	if txt := env("text.done"); txt != "" {
		config.Text.Done = txt
	}

	return config
}
