package views

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ResetURLFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)
	token := c.Param("url")
	if err := v.authSvc.CheckResetToken(c.Request().Context(), token); err != nil {
		return v.error(http.StatusBadRequest, "failed to get url for reset", err)
	}
	switch c.Request().Method {
	case http.MethodGet:
		data := struct {
			Year         int
			Context      *Context
			User         user.User
			URL          string
			VisitorCount int
		}{Year: time.Now().Year(), Context: c1, URL: token, VisitorCount: v.GetVisitorCount()}
		return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.ResetTemplate, templates.NoNavType)
	case http.MethodPost:
		data := struct {
			Error string `json:"error"`
		}{}
		if err := v.authSvc.ResetPassword(c.Request().Context(), token,
			c.FormValue("newPassword"), c.FormValue("confirmationPassword")); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}
		if err := v.clearMessagesInSession(c); err != nil {
			slog.Info(fmt.Sprintf("failed to clear messages for reset: %+v", err))
		}
		v.flash(c, c1, "successfully reset password")
		return c.JSON(http.StatusOK, data)
	default:
		return v.invalidMethodUsed(c)
	}
}

func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ResetUserPasswordFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to parse user id for reset, error: %w", err)
	}
	u, err := v.userSvc.Get(c.Request().Context(), userID)
	if err != nil {
		return fmt.Errorf("failed to get user for reset, user id: %d, error: %w", userID, err)
	}
	res, err := v.userSvc.ResetPassword(c.Request().Context(), userID)
	if err != nil {
		return fmt.Errorf("failed to reset password, user id: %d, error: %w", userID, err)
	}
	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}
	if res.EmailSent {
		message.Message = fmt.Sprintf("Reset email sent to: \"%s\"", u.Email)
	} else {
		message.Message = fmt.Sprintf("Please forward the link to this email: %s, reset link: %s", u.Email, res.ResetURL)
		message.Error = errors.New("failed to send reset email")
	}
	return c.JSON(http.StatusOK, message)
}
