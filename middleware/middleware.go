package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

// New initialises web server middleware
func New(e *echo.Echo, domainName string) {
	config := echoMw.CORSConfig{
		AllowCredentials: true,
		Skipper:          echoMw.DefaultSkipper,
		AllowOrigins: []string{
			"http://afc.bswdi.co.uk",
			"https://afc.bswdi.co.uk",
			"http://" + domainName,
			"https://" + domainName,
		},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAccessControlAllowCredentials, echo.HeaderAccessControlAllowOrigin, echo.HeaderAuthorization},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}

	e.Pre(echoMw.RemoveTrailingSlash())
	e.Use(echoMw.Logger())
	e.Use(echoMw.Recover())
	e.Use(echoMw.CORSWithConfig(config))
}
