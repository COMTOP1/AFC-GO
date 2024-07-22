package views

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	url := c.Param("url")

	id, found := v.cache.Get(url)
	if !found {
		return errors.New("failed to get url for reset")
	}

	originalUser, err := v.user.GetUser(c.Request().Context(), user.User{ID: id.(int)})
	if err != nil {
		v.cache.Delete(url)
		return fmt.Errorf("url is invalid, failed to get user: %w", err)
	}

	switch c.Request().Method {
	case "GET":
		year, _, _ := time.Now().Date()

		data := struct {
			Context *Context
			User    user.User
			URL     string
			Year    int
		}{
			Context: c1,
			User:    user.User{},
			URL:     url,
			Year:    year,
		}

		return v.template.RenderTemplate(c.Response(), data, templates.ResetTemplate, templates.NoNavType)
	case "POST":
		data := struct {
			Error string `json:"error"`
		}{}
		password := c.FormValue("newPassword")
		if password != c.FormValue("confirmationPassword") {
			data.Error = "new passwords doesn't match"
			return c.JSON(http.StatusOK, data)
		}

		log.Printf("%#v", originalUser)

		originalUser.Password = null.StringFrom(password)

		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			data.Error = fmt.Sprintf("new password doesn't meet the minimum password requirements: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		var roleDB role.Role

		roleDB, err = role.GetRole(originalUser.TempRole)
		if err != nil {
			log.Printf("failed to parse role for reset: %+v", err)
		}
		originalUser.Role = roleDB

		err = v.user.EditUserPassword(c.Request().Context(), originalUser, v.conf.Security.Iterations, v.conf.Security.KeyLength)
		if err != nil {
			log.Printf("failed to reset password: %+v", err)
			data.Error = fmt.Sprintf("failed to reset password: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		v.cache.Delete(url)
		log.Printf("updated user: %s", originalUser.Email)

		err = v.clearMessagesInSession(c)
		if err != nil {
			log.Printf("failed to clear message: %+v", err)
		}

		c1.Message = "successfully reset password"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for reset url password: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	default:
		return nil
	}
}

func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse userid for reset: %w", err)
		}

		userDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for reset: %w", err)
		}

		userDB.ResetPassword = true

		var roleDB role.Role

		roleDB, err = role.GetRole(userDB.TempRole)
		if err != nil {
			log.Printf("failed to parse role for reset: %+v", err)
		}
		userDB.Role = roleDB

		_, err = v.user.EditUser(c.Request().Context(), userDB)
		if err != nil {
			return fmt.Errorf("failed to update user for reset: %w", err)
		}

		url := uuid.NewString()
		v.cache.Set(url, userDB.ID, 7*24*time.Hour)

		var message struct {
			Message string `json:"message"`
			Error   error  `json:"error"`
		}

		mailer := v.mailer.ConnectMailer()

		// Valid request, send email with reset code
		if mailer != nil {
			var emailTemplate *template.Template
			emailTemplate, err = v.template.GetEmailTemplate(templates.ResetEmailTemplate)
			if err != nil {
				return fmt.Errorf("failed to render email for reset: %w", err)
			}

			file := mail.Mail{
				Subject: "AFC Security - Reset Password",
				Tpl:     emailTemplate,
				To:      userDB.Email,
				From:    "AFC Security <no-reply.afc@bswdi.co.uk>",
				TplData: struct {
					Email string
					URL   string
				}{
					Email: userDB.Email,
					URL:   fmt.Sprintf("https://%s/reset/%s", v.conf.DomainName, url),
				},
			}

			err = mailer.SendMail(file)
			if err != nil {
				message.Message = fmt.Sprintf("Please forward the link to this email: %s, reset link: https://%s/reset/%s", userDB.Email, v.conf.DomainName, url)
				message.Error = fmt.Errorf("failed to send mail: %w", err)
				log.Printf("failed to send mail: %+v", err)
				log.Printf("password reset requested for email: %s by user: %d", userDB.Email, c1.User.ID)
				return c.JSON(http.StatusOK, message)
			}
			_ = mailer.Close()

			log.Printf("password reset requested for email: %s by user: %d", userDB.Email, c1.User.ID)
			message.Message = fmt.Sprintf("Reset email sent to: \"%s\"", userDB.Email)
		} else {
			message.Message = fmt.Sprintf("No mailer present\nPlease forward the link to this email: %s, reset link: https://%s/reset/%s", userDB.Email, v.conf.DomainName, url)
			message.Error = errors.New("no mailer present")
			log.Printf("no Mailer present")
			log.Printf("password reset requested for email: %s by user: %d", userDB.Email, c1.User.ID)
		}
		log.Printf("reset for %d (%s) requested by %d (%s)", userDB.ID, userDB.Name, c1.User.ID, c1.User.Name)

		return c.JSON(http.StatusOK, message)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}
