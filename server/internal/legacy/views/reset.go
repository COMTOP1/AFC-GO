package views

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ResetURLFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	url := c.Param("url")

	id, found := v.GetResetToken(c.Request().Context(), url)
	if !found {
		return v.error(http.StatusBadRequest, "failed to get url for reset",
			fmt.Errorf("failed to get url for reset, url: %s", url))
	}

	originalUser, err := v.user.GetUser(c.Request().Context(), user.User{ID: id})
	if err != nil {
		v.DeleteResetToken(c.Request().Context(), url)
		return v.error(http.StatusInternalServerError, "failed to get user for reset",
			fmt.Errorf("url is invalid, failed to get user, error: %w", err))
	}

	switch c.Request().Method {
	case "GET":
		year, _, _ := time.Now().Date()

		data := struct {
			Context      *Context
			User         user.User
			URL          string
			Year         int
			VisitorCount int
		}{
			Context:      c1,
			User:         user.User{},
			URL:          url,
			Year:         year,
			VisitorCount: v.GetVisitorCount(),
		}

		return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.ResetTemplate, templates.NoNavType)
	case "POST":
		data := struct {
			Error string `json:"error"`
		}{}
		password := c.FormValue("newPassword")
		if password != c.FormValue("confirmationPassword") {
			data.Error = "new passwords doesn't match"
			return c.JSON(http.StatusOK, data)
		}

		originalUser.Password = null.StringFrom(password)

		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			data.Error = fmt.Sprintf("new password doesn't meet the minimum password requirements: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		err = v.user.EditUserPassword(c.Request().Context(), originalUser, v.conf.Security.ScryptWorkFactor,
			v.conf.Security.ScryptBlockSize, v.conf.Security.ScryptParallelismFactor, v.conf.Security.KeyLength)
		if err != nil {
			slog.Info(fmt.Sprintf("failed to reset password, error: %+v", err))
			data.Error = fmt.Sprintf("failed to reset password: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		v.DeleteResetToken(c.Request().Context(), url)
		slog.Info("updated user password: " + originalUser.Email)

		err = v.clearMessagesInSession(c)
		if err != nil {
			slog.Info(fmt.Sprintf("failed to clear message for reset, error: %+v", err))
		}

		c1.Message = "successfully reset password"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			slog.Info(fmt.Sprintf("failed to set data for reset url password, error: %+v", err))
		}

		return c.JSON(http.StatusOK, data)
	default:
		return nil
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
