package views

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (v *Views) Account(c echo.Context) error {
	return c.JSON(http.StatusOK, "account")
}
