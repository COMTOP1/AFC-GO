package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const swaggerIndex = "/api/v1/swagger/index.html"

// NewAPI mounts /api and returns the /api/v1 group, protected by CSRF.
// Unknown /api paths are JSON 404s rather than the web client's not-found page.
func NewAPI(e *echo.Echo, secureCookies bool) *echo.Group {
	api := e.Group("/api")
	api.GET("", swaggerRedirect)
	api.GET("/health", Health) // legacy alias, kept for existing monitors
	api.RouteNotFound("/*", func(echo.Context) error { return echo.ErrNotFound })

	v1 := api.Group("/v1", CSRF(secureCookies))
	v1.GET("", swaggerRedirect)
	v1.GET("/health", Health)
	v1.GET("/swagger/*", echoSwagger.WrapHandler)
	return v1
}

func swaggerRedirect(c echo.Context) error {
	return c.Redirect(http.StatusFound, swaggerIndex)
}

// HealthResponse is the body of GET /health.
type HealthResponse struct {
	Status string `json:"status"`
}

// Health reports that the server is up.
//
//	@Summary	Health check
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	web.HealthResponse
//	@Router		/health [get]
func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
