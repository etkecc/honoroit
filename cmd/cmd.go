package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	zlogsentry "github.com/archdx/zerolog-sentry"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mileusna/crontab"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
	"gitlab.com/etke.cc/go/healthchecks/v2"
	"gitlab.com/etke.cc/go/psd"
	"gitlab.com/etke.cc/linkpearl"

	"gitlab.com/etke.cc/honoroit/config"
	"gitlab.com/etke.cc/honoroit/controllers"
	"gitlab.com/etke.cc/honoroit/matrix"
	mxconfig "gitlab.com/etke.cc/honoroit/matrix/config"
	"gitlab.com/etke.cc/honoroit/metrics"
)

var (
	e    *echo.Echo
	hc   *healthchecks.Client
	bot  *matrix.Bot
	ctab *crontab.Crontab
	log  zerolog.Logger
)

func main() {
	cfg := config.New()
	initLog(cfg)
	initHTTP(cfg)
	initHealthchecks(cfg)
	metrics.InitMetrics()
	defer recovery()

	log.Info().Msg("#############################")
	log.Info().Msg("Honoroit")
	log.Info().Msg("#############################")

	if err := initBot(cfg); err != nil {
		log.Error().Err(err).Msg("cannot initialize the bot")
		return
	}
	ctab = crontab.New()
	initShutdown()

	if err := ctab.AddJob("0 15 * * *", bot.AutoCloseRequests); err != nil {
		log.Error().Err(err).Msg("cannot add cron job")
		return
	}

	go e.Start(cfg.Port) //nolint:errcheck // nobody cares

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
	sentryWriter, err := zlogsentry.New(cfg.Monitoring.SentryDSN, zlogsentry.WithBreadcrumbs())
	if err == nil {
		w = io.MultiWriter(sentryWriter, consoleWriter)
	} else {
		w = consoleWriter
	}
	log = zerolog.New(w).With().Timestamp().Caller().Logger()
}

func initHealthchecks(cfg *config.Config) {
	if cfg.Monitoring.HealthchecksUUID == "" {
		return
	}
	hc = healthchecks.New(
		healthchecks.WithBaseURL(cfg.Monitoring.HealthchecksURL),
		healthchecks.WithCheckUUID(cfg.Monitoring.HealthchecksUUID),
		healthchecks.WithErrLog(func(operation string, err error) {
			log.Error().Err(err).Str("op", operation).Msg("healthchecks operation failed")
		}),
	)
	hc.Start(strings.NewReader("starting honoroit"))
	go hc.Auto(cfg.Monitoring.HealthchecksDuration)
}

func initHTTP(cfg *config.Config) {
	e = echo.New()
	e.Logger = lecho.From(log)
	controllers.ConfigureRouter(e, cfg.Auth.Metrics)
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
	psdc := psd.NewClient(cfg.Auth.PSD.URL, cfg.Auth.PSD.Login, cfg.Auth.PSD.Password)
	mxc := mxconfig.New(lp)
	bot, err = matrix.NewBot(lp, &log, mxc, psdc, cfg.Prefix, cfg.RoomID, cfg.CacheSize)
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
	ctab.Shutdown()

	if hc != nil {
		hc.Shutdown()
		hc.ExitStatus(0, strings.NewReader("shutting down honoroit"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	e.Shutdown(ctx) //nolint:errcheck // nobody cares
}

func recovery() {
	r := recover()
	// no problem just shutdown
	if r == nil {
		return
	}

	if hc != nil {
		hc.ExitStatus(1, strings.NewReader(fmt.Sprintf("panic: %v", r)))
	}
	err, ok := r.(error)
	if !ok {
		log.Error().Interface("panic", r).Msg("panic")
		return
	}
	log.Error().Err(err).Msg("panic")
}
