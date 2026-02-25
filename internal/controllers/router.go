package controllers

import (
	"bytes"
	"net/http"

	echobasicauth "github.com/etkecc/go-echo-basic-auth"
	"github.com/etkecc/go-kit"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/etkecc/honoroit/internal/metrics"
)

// ConfigureRouter configures echo router
func ConfigureRouter(e *echo.Echo, auth *echobasicauth.Auth) {
	//nolint:staticcheck // new logger is less convenient
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/_health"
		},
		Format:           `${custom} - - [${time_custom}] "${method} ${path} ${protocol}" ${status} ${bytes_out} "${referer}" "${user_agent}"` + "\n",
		CustomTimeFormat: "2/Jan/2006:15:04:05 -0700",
		CustomTagFunc: func(c echo.Context, w *bytes.Buffer) (int, error) {
			return w.WriteString(kit.AnonymizeIP(c.RealIP()))
		},
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
