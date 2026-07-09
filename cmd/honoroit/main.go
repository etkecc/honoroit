package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/etkecc/go-healthchecks/v2"
	"github.com/etkecc/go-kit/crontab"
	"github.com/etkecc/go-linkpearl"
	"github.com/etkecc/go-redmine"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
	_ "modernc.org/sqlite"

	"github.com/etkecc/honoroit/internal/config"
	"github.com/etkecc/honoroit/internal/controllers"
	"github.com/etkecc/honoroit/internal/matrix"
	mxconfig "github.com/etkecc/honoroit/internal/matrix/config"
	"github.com/etkecc/honoroit/internal/metrics"
)

var (
	e    *echo.Echo
	hc   *healthchecks.Client
	rdm  *redmine.Redmine
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

	var err error
	rdm, err = redmine.New(
		redmine.WithLog(&log),
		redmine.WithHost(cfg.Redmine.Host),
		redmine.WithAPIKey(cfg.Redmine.APIKey),
		redmine.WithProjectIdentifier(cfg.Redmine.ProjectID),
		redmine.WithTrackerID(cfg.Redmine.TrackerID),
		redmine.WithWaitingForOperatorStatusID(cfg.Redmine.NewStatus),
		redmine.WithWaitingForCustomerStatusID(cfg.Redmine.InProgressStatus),
		redmine.WithDoneStatusID(cfg.Redmine.DoneStatus),
	)
	if err != nil {
		log.Warn().Err(err).Msg("cannot initialize redmine")
	}
	if rdm.Enabled() {
		log.Info().Msg("redmine integration enabled")
	}

	if err := initBot(cfg, rdm); err != nil {
		log.Error().Err(err).Msg("cannot initialize the bot")
		return
	}
	ctab = crontab.New(crontab.WithPanicHandler(func(spec string, recovered any) {
		log.Error().Str("spec", spec).Any("panic", recovered).Msg("cron job panic")
	}))
	initShutdown()

	ctab.MustAddJob("0 15 * * *", bot.AutoCloseRequests)
	ctab.MustAddJob("* * * * *", bot.SyncIssues)

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
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, PartsExclude: []string{zerolog.TimestampFieldName}}
	log = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
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

func initBot(cfg *config.Config, rdm *redmine.Redmine) error {
	if cfg.DB.Dialect == "sqlite3" {
		cfg.DB.Dialect = "sqlite"
	}
	db, err := sql.Open(cfg.DB.Dialect, cfg.DB.DSN)
	if err != nil {
		return err
	}
	// workaround for sqlite's SQLITE_BUSY
	if cfg.DB.Dialect == "sqlite" {
		db.SetMaxOpenConns(1)
	}
	lp, err := linkpearl.New(&linkpearl.Config{
		Homeserver:        cfg.Homeserver,
		Login:             cfg.Login,
		Password:          cfg.Password,
		SharedSecret:      cfg.SharedSecret,
		Token:             cfg.Token,
		DB:                db,
		Dialect:           cfg.DB.Dialect,
		AccountDataSecret: cfg.DataSecret,
		Logger:            log,
	})
	if err != nil {
		return err
	}
	mxc := mxconfig.New(lp)
	bot, err = matrix.NewBot(lp, &log, mxc, rdm, cfg.Prefix, cfg.RoomID, cfg.CacheSize, cfg.NoEncryptionWarning)
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
	// drain cron before stopping the bot: AutoCloseRequests/SyncIssues call into the matrix client bot.Stop() tears down
	ctabCtx, ctabCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctabCancel()
	if err := ctab.Shutdown(ctabCtx); err != nil {
		log.Warn().Err(err).Msg("cron shutdown did not drain cleanly")
	}
	bot.Stop()
	rdm.Shutdown()

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
