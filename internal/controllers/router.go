package controllers

import (
	"fmt"
	"net/http"
	"time"

	apm "github.com/etkecc/go-apm/v2"
	echobasicauth "github.com/etkecc/go-echo-basic-auth/v2"
	"github.com/etkecc/go-kit"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/etkecc/honoroit/internal/metrics"
)

const logTimeFormat = "2/Jan/2006:15:04:05 -0700"

// ConfigureRouter configures echo router
func ConfigureRouter(e *echo.Echo, auth *echobasicauth.Auth) {
	e.Use(apm.WithSentry())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c *echo.Context) bool {
			return c.Request().URL.Path == "/_health"
		},
		LogMethod:       true,
		LogURIPath:      true,
		LogProtocol:     true,
		LogRemoteIP:     true,
		LogReferer:      true,
		LogUserAgent:    true,
		LogStatus:       true,
		LogResponseSize: true,
		HandleError:     true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status >= http.StatusInternalServerError {
				c.Logger().Error("request failed", "status", v.Status, "path", v.URIPath, "error", v.Error)
			}
			c.Logger().Info(fmt.Sprintf(`%s - - [%s] %q %d %d %q %q`,
				kit.AnonymizeIP(v.RemoteIP),
				time.Now().Format(logTimeFormat),
				v.Method+" "+v.URIPath+" "+v.Protocol,
				v.Status,
				max(v.ResponseSize, 0),
				v.Referer,
				v.UserAgent,
			))
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set(echo.HeaderReferrerPolicy, "origin")
			return next(c)
		}
	})
	// XFF trusted from loopback, link-local and private peers; keep aligned with the proxy topology
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(true),
		echo.TrustPrivateNet(true),
	)

	e.GET("/_health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", echo.WrapHandler(&metrics.Handler{}), echobasicauth.NewMiddleware(auth))
}
