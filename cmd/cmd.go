package main

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	zlogsentry "github.com/archdx/zerolog-sentry"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"gitlab.com/etke.cc/go/healthchecks"
	"gitlab.com/etke.cc/linkpearl"

	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/matrix"
	mxconfig "gitlab.com/etke.cc/honoroit/matrix/config"
	"gitlab.com/etke.cc/honoroit/metrics"
)

var (
	hc  *healthchecks.Client
	bot *matrix.Bot
	srv *http.Server
	log zerolog.Logger
)

func main() {
	cfg := config.New()
	initLog(cfg)
	initHealthchecks(cfg)
	metrics.InitMetrics()
	go initHTTP(cfg)
	defer recovery()

	log.Info().Msg("#############################")
	log.Info().Msg("Honoroit")
	log.Info().Msg("#############################")

	if err := initBot(cfg); err != nil {
		log.Error().Err(err).Msg("cannot initialize the bot")
		return
	}
	initShutdown()

	log.Debug().Msg("starting bot...")
	if err := bot.Start(); err != nil {
		log.Error().Err(err).Msg("matrix bot crashed")
	}
}

func initLog(cfg *config.Config) {
	loglevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		loglevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(loglevel)
	var w io.Writer
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, PartsExclude: []string{zerolog.TimestampFieldName}}
	sentryWriter, err := zlogsentry.New(cfg.Monitoring.SentryDSN)
	if err == nil {
		w = io.MultiWriter(sentryWriter, consoleWriter)
	} else {
		w = consoleWriter
	}
	log = zerolog.New(w).With().Timestamp().Caller().Logger()
}

func initHealthchecks(cfg *config.Config) {
	if cfg.Monitoring.HealchecksUUID == "" {
		return
	}
	hc = healthchecks.New(cfg.Monitoring.HealchecksUUID, func(operation string, err error) {
		log.Error().Err(err).Str("op", operation).Msg("healthchecks operation failed")
	})
	hc.Start(strings.NewReader("starting honoroit"))
	go hc.Auto(cfg.Monitoring.HealthechsDuration)
}

func initHTTP(cfg *config.Config) {
	srv = &http.Server{Addr: ":" + cfg.Port, Handler: nil}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("http server failed")
	}
}

func initBot(cfg *config.Config) error {
	db, err := sql.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		return err
	}
	lp, err := linkpearl.New(&linkpearl.Config{
		Homeserver:        cfg.Homeserver,
		Login:             cfg.Login,
		Password:          cfg.Password,
		SharedSecret:      cfg.SharedSecret,
		DB:                db,
		Dialect:           cfg.DB.Dialect,
		AccountDataSecret: cfg.DataSecret,
		Logger:            log,
	})
	if err != nil {
		return err
	}
	mxc := mxconfig.New(lp)
	bot, err = matrix.NewBot(lp, &log, mxc, cfg.Prefix, cfg.RoomID, cfg.CacheSize)
	return err
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

func recovery() {
	err := recover()
	// no problem just shutdown
	if err == nil {
		return
	}

	log.Error().Err(err.(error)).Msg("panic")
}
