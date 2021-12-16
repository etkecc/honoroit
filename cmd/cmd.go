package main

import (
	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/logger"
	"gitlab.com/etke.cc/honoroit/matrix"
)

var version = "development"

func main() {
	cfg := config.New()
	log := logger.New("honoroit.", cfg.LogLevel)

	log.Info("#############################")
	log.Info("Honoroit " + version)
	log.Info("Matrix: true")
	log.Info("#############################")

	botConfig := &matrix.Config{
		Homeserver: cfg.Homeserver,
		Login:      cfg.Login,
		Password:   cfg.Password,
		Token:      cfg.Token,
		LogLevel:   cfg.LogLevel,
		RoomID:     cfg.RoomID,
		Text:       (*matrix.Text)(&cfg.Text),
	}
	bot, err := matrix.NewBot(botConfig)
	if err != nil {
		log.Error("cannot create the matrix bot: %v", err)
		return
	}
	defer bot.Stop()
	log.Debug("bot has been created")

	if err = bot.WithStore(); err != nil {
		log.Error("cannot initialize data store: %v", err)
		return
	}
	log.Debug("data store initialized")

	log.Debug("starting bot...")
	if err = bot.Start(); err != nil {
		log.Error("cannot start the matrix bot: %v", err)
	}
}
