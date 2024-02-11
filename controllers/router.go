package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echobasicauth "gitlab.com/etke.cc/go/echo-basic-auth"

	"gitlab.com/etke.cc/honoroit/metrics"
)

// ConfigureRouter configures echo router
func ConfigureRouter(e *echo.Echo, auth *echobasicauth.Auth) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/_health"
		},
		Format:           `${remote_ip} - - [${time_custom}] "${method} ${path} ${protocol}" ${status} ${bytes_out} "${referer}" "${user_agent}"` + "\n",
		CustomTimeFormat: "2/Jan/2006:15:04:05 -0700",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderReferrerPolicy, "origin")
			return next(c)
		}
	})
	e.HideBanner = true
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLoopback(true),
		echo.TrustLinkLocal(true),
		echo.TrustPrivateNet(true),
	)

	e.GET("/_health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/metrics", echo.WrapHandler(&metrics.Handler{}), echobasicauth.NewMiddleware(auth))
}
