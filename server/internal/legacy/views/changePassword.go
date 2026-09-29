package views

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ChangePasswordFunc handles the password change from a user
func (v *Views) ChangePasswordFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ChangePasswordFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)
		data := struct {
			Error string `json:"error"`
		}{}
		if err := v.authSvc.ChangePassword(c.Request().Context(), c1.User,
			c.FormValue("oldPassword"), c.FormValue("newPassword"), c.FormValue("confirmationPassword")); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}
		v.flash(c, c1, "successfully changed password")
		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}
