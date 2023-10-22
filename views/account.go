package views

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func (v *Views) Account(c echo.Context) error {
	return c.JSON(http.StatusOK, "account")
}
