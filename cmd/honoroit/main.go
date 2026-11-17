package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	apm "github.com/etkecc/go-apm/v2"
	"github.com/etkecc/go-healthchecks/v2"
	"github.com/etkecc/go-kit/crontab"
	"github.com/etkecc/go-linkpearl"
	"github.com/etkecc/go-redmine"
	"github.com/labstack/echo/v5"
	_ "github.com/lib/pq"
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
	log  *slog.Logger

	shutdownSignals = []os.Signal{
		os.Interrupt, syscall.SIGABRT, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM,
	}
)

const (
	banner          = "#############################"
	shutdownTimeout = 5 * time.Second
)

func main() {
	os.Exit(run())
}

// run starts the bot, serves until a signal or a failure, and returns the process exit code
func run() (code int) {
	cfg := config.New()
	log = initLog(cfg)
	metrics.InitMetrics()
	defer apm.Flush()

	ctx, stop := signal.NotifyContext(apm.NewContext(), shutdownSignals...)
	defer stop()
	serveCtx, cancelServe := context.WithCancel(ctx)
	defer cancelServe()

	defer func() {
		if r := recover(); r != nil {
			log.Error("panic", "error", fmt.Sprintf("%+v", r))
			if hc != nil {
				hc.ExitStatus(1, strings.NewReader(fmt.Sprintf("panic: %v", r)))
				hc.Shutdown()
			}
			code = 1
		}
	}()

	if err := initHTTP(cfg); err != nil {
		log.Error("cannot initialize the HTTP server", "error", err)
		stopServices(1, "startup failed", true)
		return 1
	}
	initHealthchecks(cfg)

	log.Info(banner)
	log.Info("Honoroit")
	log.Info(banner)

	var err error
	rdm, err = redmine.New(
		redmine.WithLog(log),
		redmine.WithHost(cfg.Redmine.Host),
		redmine.WithAPIKey(cfg.Redmine.APIKey),
		redmine.WithProjectIdentifier(cfg.Redmine.ProjectID),
		redmine.WithTrackerID(cfg.Redmine.TrackerID),
		redmine.WithWaitingForOperatorStatusID(cfg.Redmine.NewStatus),
		redmine.WithWaitingForCustomerStatusID(cfg.Redmine.InProgressStatus),
		redmine.WithDoneStatusID(cfg.Redmine.DoneStatus),
	)
	if err != nil {
		log.Warn("cannot initialize redmine", "error", err)
	}
	if rdm.Enabled() {
		log.Info("redmine integration enabled")
	}

	if err := initBot(cfg, rdm); err != nil {
		log.Error("cannot initialize the bot", "error", err)
		stopServices(1, "startup failed", true)
		return 1
	}
	ctab = crontab.New(crontab.WithPanicHandler(func(spec string, recovered any) {
		log.Error("cron job panic", "spec", spec, "panic", recovered)
	}))
	ctab.MustAddJob("0 15 * * *", bot.AutoCloseRequests)
	ctab.MustAddJob("* * * * *", bot.SyncIssues)

	serveErr := make(chan error, 1)
	botErr := make(chan error, 1)
	go func() {
		serveErr <- echo.StartConfig{
			Address:         listenAddress(cfg.Port),
			HideBanner:      true,
			GracefulTimeout: shutdownTimeout,
		}.Start(serveCtx, e)
	}()

	log.Debug("starting bot...")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				botErr <- fmt.Errorf("matrix bot panic: %+v", r)
			}
		}()
		botErr <- bot.Start()
	}()

	code = awaitShutdown(ctx, cancelServe, serveErr, botErr)
	log.Info("stopped")
	return code
}

func listenAddress(port string) string {
	addr := strings.TrimSpace(port)
	if strings.Contains(addr, ":") {
		return addr
	}
	return net.JoinHostPort("", addr)
}

func initLog(cfg *config.Config) *slog.Logger {
	apm.SetName("honoroit")
	apm.SetLogLevel(cfg.LogLevel)
	apm.SetSentryDSN(cfg.Monitoring.SentryDSN)
	logger := apm.Log()
	slog.SetDefault(logger)
	return logger
}

func initHTTP(cfg *config.Config) error {
	if err := cfg.Auth.Metrics.Validate(); err != nil {
		return fmt.Errorf("invalid metrics auth: %w", err)
	}
	e = echo.New()
	e.Logger = log
	controllers.ConfigureRouter(e, cfg.Auth.Metrics)
	return nil
}

func initHealthchecks(cfg *config.Config) {
	if cfg.Monitoring.HealthchecksUUID == "" {
		return
	}
	hc = healthchecks.New(
		healthchecks.WithBaseURL(cfg.Monitoring.HealthchecksURL),
		healthchecks.WithCheckUUID(cfg.Monitoring.HealthchecksUUID),
		healthchecks.WithErrLog(func(operation string, err error) {
			log.Warn("healthchecks operation failed", "op", operation, "error", err)
		}),
	)
	apm.SetHealthchecks(hc)
	hc.Start(strings.NewReader("starting honoroit"))
	go hc.Auto(cfg.Monitoring.HealthchecksDuration)
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
	})
	if err != nil {
		return err
	}
	mxc := mxconfig.New(lp)
	bot, err = matrix.NewBot(lp, log, mxc, rdm, cfg.Prefix, cfg.RoomID, cfg.CacheSize, cfg.NoEncryptionWarning)
	return err
}

func awaitShutdown(ctx context.Context, cancelServe context.CancelFunc, serveErr, botErr <-chan error) int {
	code := 0
	serveDone, botDone := false, false
	select {
	case <-ctx.Done():
		log.Info("shutting down", "reason", context.Cause(ctx))
	case err := <-serveErr:
		serveDone = true
		if err != nil {
			log.Error("http server failed", "error", err)
			code = 1
		}
	case err := <-botErr:
		botDone = true
		log.Error("matrix bot stopped unexpectedly", "error", err)
		code = 1
	}

	cancelServe()
	if !serveDone {
		if err := <-serveErr; err != nil {
			log.Error("http server shutdown failed", "error", err)
			code = 1
		}
	}
	stopServices(code, "shutting down honoroit", !botDone)
	return code
}

func stopServices(exitCode int, reason string, drainBot bool) {
	if ctab != nil {
		ctabCtx, ctabCancel := context.WithTimeout(apm.NewContext(), shutdownTimeout)
		if err := ctab.Shutdown(ctabCtx); err != nil {
			log.Warn("cron shutdown did not drain cleanly", "error", err)
		}
		ctabCancel()
	}
	if drainBot && bot != nil {
		bot.Stop()
	}
	if rdm != nil {
		rdm.Shutdown()
	}
	if hc != nil {
		hc.ExitStatus(exitCode, strings.NewReader(reason))
		hc.Shutdown()
	}
}
