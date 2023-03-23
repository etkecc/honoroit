package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/etke.cc/go/healthchecks"
	"gitlab.com/etke.cc/go/logger"
	"maunium.net/go/mautrix/id"

	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/matrix"
	"gitlab.com/etke.cc/honoroit/metrics"
)

var (
	version = "development"
	hc      *healthchecks.Client
	bot     *matrix.Bot
	srv     *http.Server
	log     *logger.Logger
)

func main() {
	var err error
	cfg := config.New()
	log = logger.New("honoroit.", cfg.LogLevel)
	initSentry(cfg)
	initHealthchecks(cfg)
	metrics.InitMetrics()
	go initHTTP(cfg)
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
		Dsn:              cfg.Monitoring.SentryDSN,
		Release:          "honoroit@" + version,
		Environment:      env,
		AttachStacktrace: true,
		TracesSampleRate: float64(cfg.Monitoring.SentrySampleRate) / 100,
	})
	if err != nil {
		log.Fatal("cannot initialize sentry: %v", err)
	}
}

func initHealthchecks(cfg *config.Config) {
	if cfg.Monitoring.HealchecksUUID == "" {
		return
	}
	hc = healthchecks.New(cfg.Monitoring.HealchecksUUID, func(operation string, err error) {
		log.Error("healthchecks operation %q failed: %v", operation, err)
	})
	hc.Start(strings.NewReader("starting honoroit"))
	go hc.Auto(cfg.Monitoring.HealthechsDuration)
}

func initHTTP(cfg *config.Config) {
	srv = &http.Server{Addr: ":" + cfg.Port, Handler: nil}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("http server failed: %v", err)
	}
}

func initBot(cfg *config.Config) {
	db, err := sql.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		log.Fatal("cannot initialize SQL database: %v", err)
	}
	botConfig := &matrix.Config{
		Homeserver:     cfg.Homeserver,
		Login:          cfg.Login,
		Password:       cfg.Password,
		LogLevel:       cfg.LogLevel,
		Prefix:         cfg.Prefix,
		RoomID:         cfg.RoomID,
		AllowedUsers:   cfg.AllowedUsers,
		IgnoredRooms:   cfg.IgnoredRooms,
		IgnoreNoThread: cfg.IgnoreNoThread,
		Text:           (*matrix.Text)(&cfg.Text),
		DB:             db,
		Dialect:        cfg.DB.Dialect,
		CacheSize:      cfg.CacheSize,
		NoEncryption:   cfg.NoEncryption,
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
			shutdown()
			os.Exit(0)
		}
	}()
}

func shutdown() {
	bot.Stop()

	if hc != nil {
		hc.Shutdown()
		hc.ExitStatus(0, strings.NewReader("shutting down honoroit"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx) //nolint:errcheck // nobody cares
}

func recovery(roomID string) {
	defer sentry.Flush(5 * time.Second)
	err := recover()
	// no problem just shutdown
	if err == nil {
		return
	}

	// try to send that error to matrix and log, if available
	if bot != nil {
		bot.Error(id.RoomID(roomID), nil, sentry.CurrentHub(), "recovery(): %v", err)
	}

	sentry.CurrentHub().Recover(err)
	sentry.Flush(5 * time.Second)
}
