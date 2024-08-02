package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"
)

// ChangePasswordFunc handles the password change from a user
func (v *Views) ChangePasswordFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		oldPassword := c.FormValue("oldPassword")

		data := struct {
			Error string `json:"error"`
		}{}

		c1.User.Password = null.StringFrom(oldPassword)

		_, _, err := v.user.VerifyUser(c.Request().Context(), c1.User, v.conf.Security.Iterations, v.conf.Security.ScryptWorkFactor, v.conf.Security.ScryptBlockSize, v.conf.Security.ScryptParallelismFactor, v.conf.Security.KeyLength)
		if err != nil {
			data.Error = "old password is not correct"
			return c.JSON(http.StatusOK, data)
		}

		password := c.FormValue("newPassword")

		if password != c.FormValue("confirmationPassword") {
			data.Error = "new passwords doesn't match"
			return c.JSON(http.StatusOK, data)
		}

		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			data.Error = "new password doesn't meet the old requirements: " + errString
			return c.JSON(http.StatusOK, data)
		}

		c1.User.Password = null.StringFrom(password)

		err = v.user.EditUserPassword(c.Request().Context(), c1.User, v.conf.Security.ScryptWorkFactor,
			v.conf.Security.ScryptBlockSize, v.conf.Security.ScryptParallelismFactor, v.conf.Security.KeyLength)
		if err != nil {
			log.Printf("failed to change password: %+v", err)
			data.Error = fmt.Sprintf("failed to change password: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully changed password"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for change password: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}
