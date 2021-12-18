package main

import (
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/logger"
	"gitlab.com/etke.cc/honoroit/matrix"
)

const fatalmessage = "recovery(): %v"

var (
	version = "development"
	bot     *matrix.Bot
	log     *logger.Logger
)

func main() {
	var err error
	cfg := config.New()
	log = logger.New("honoroit.", cfg.LogLevel)
	defer recovery(cfg.RoomID)

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
	bot, err = matrix.NewBot(botConfig)
	if err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot create the matrix bot: %v", err)
	}
	defer bot.Stop()
	log.Debug("bot has been created")

	if err = bot.WithStore(); err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot initialize data store: %v", err)
	}
	log.Debug("data store initialized")

	log.Debug("starting bot...")
	if err = bot.Start(); err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot start the matrix bot: %v", err)
	}
}

func recovery(roomID string) {
	err := recover()
	// no problem just shutdown
	if err == nil {
		return
	}

	// try to send that error to matrix and log, if available
	if bot != nil {
		bot.Error(id.RoomID(roomID), fatalmessage, err)
		return
	}

	log.Error(fatalmessage, err)
}
