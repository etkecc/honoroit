package main

import (
	"database/sql"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.sr.ht/~xn/cache"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/logger"
	"gitlab.com/etke.cc/honoroit/matrix"
)

var (
	version = "development"
	bot     *matrix.Bot
	log     *logger.Logger
)

func main() {
	var err error
	cfg := config.New()
	log = logger.New("honoroit.", cfg.LogLevel)
	initSentry(cfg)
	defer recovery(cfg.RoomID)

	log.Info("#############################")
	log.Info("Honoroit " + version)
	log.Info("Matrix: true")
	log.Info("#############################")

	initBot(cfg)
	initShutdown()

	log.Debug("starting bot...")
	if err = bot.Start(); err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("matrix bot crashed: %v", err)
	}
}

func initSentry(cfg *config.Config) {
	env := version
	if env != "development" {
		env = "production"
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.Sentry,
		Release:     "honoroit@" + version,
		Environment: env,
	})
	if err != nil {
		log.Fatal("cannot initialize sentry: %v", err)
	}
}

func initBot(cfg *config.Config) {
	db, err := sql.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		log.Fatal("cannot initialize SQL database: %v", err)
	}
	tlru := cache.NewTLRU(2000, 1*time.Hour)
	botConfig := &matrix.Config{
		Homeserver: cfg.Homeserver,
		Login:      cfg.Login,
		Password:   cfg.Password,
		LogLevel:   cfg.LogLevel,
		Prefix:     cfg.Prefix,
		RoomID:     cfg.RoomID,
		Text:       (*matrix.Text)(&cfg.Text),
		DB:         db,
		Dialect:    cfg.DB.Dialect,
		Cache:      tlru,
	}
	bot, err = matrix.NewBot(botConfig)
	if err != nil {
		// nolint // Fatal = panic, not os.Exit()
		log.Fatal("cannot create the matrix bot: %v", err)
	}
	log.Debug("bot has been created")
}

func initShutdown() {
	listener := make(chan os.Signal, 1)
	signal.Notify(listener, os.Interrupt, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for range listener {
			bot.Stop()
			os.Exit(0)
		}
	}()
}

func recovery(roomID string) {
	defer sentry.Flush(2 * time.Second)
	err := recover()
	// no problem just shutdown
	if err == nil {
		return
	}

	// try to send that error to matrix and log, if available
	if bot != nil {
		bot.Error(id.RoomID(roomID), sentry.CurrentHub(), "recovery(): %v", err)
	}

	sentry.CurrentHub().Recover(err)
	sentry.Flush(5 * time.Second)
}
