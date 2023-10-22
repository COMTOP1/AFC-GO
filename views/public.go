package views

import (
	"github.com/labstack/echo/v4"
)

func (v *Views) Public(c echo.Context) error {
	return c.Inline("public/"+c.Param("file"), c.Param("file"))
}
