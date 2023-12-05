package views

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/infrastructure/mail"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (v *Views) UsersFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var u1 []user.User
	var err error

	u1, err = v.user.GetUsers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year  int
		Users []UserTemplate
		User  user.User
	}{
		Year:  year,
		Users: DBUsersToTemplateFormat(u1),
		User:  c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.UsersTemplate, templates.RegularType)
}

func (v *Views) UserAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		verifier := emailverifier.NewVerifier()

		var data struct {
			Error string `json:"error"`
		}

		name := c.FormValue("name")
		email := c.FormValue("email")
		phone := c.FormValue("phone")

		var message struct {
			Message string `json:"message"`
			Error   error  `json:"error"`
		}

		formRole, err := role.GetRole(c.FormValue("role"))
		if err != nil {
			log.Printf("failed to get role for userAdd: %+v", err)
			message.Error = fmt.Errorf("failed to get role for userAdd: %w", err)
			return c.JSON(http.StatusOK, message)
		}

		teamID, err := strconv.Atoi(c.FormValue("teamID"))
		if err != nil {
			log.Printf("failed to get teamID for userAdd: %+v, proceeding with no team", err)
			teamID = 0
		}
		if teamID < 0 {
			log.Println("failed to parse negative number, proceeding with no team")
			teamID = 0
		}

		password, err := utils.GenerateRandom(utils.GeneratePassword)
		if err != nil {
			return fmt.Errorf("error generating password: %w", err)
		}

		salt, err := utils.GenerateRandom(utils.GenerateSalt)
		if err != nil {
			return fmt.Errorf("error generating salt: %w", err)
		}

		hash := utils.HashPass([]byte(password), []byte(salt), v.conf.Security.Iterations, v.conf.Security.KeyLength)

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for userAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for userAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for userAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for userAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		u := user.User{
			ID:            0,
			Name:          name,
			Email:         email,
			Phone:         phone,
			TeamID:        null.IntFrom(int64(teamID)),
			Role:          formRole,
			FileName:      null.StringFrom(""),
			ResetPassword: true,
			Hash:          null.StringFrom(hex.EncodeToString(hash)),
			Salt:          null.StringFrom(salt),
		}

		_, err = v.user.AddUser(c.Request().Context(), u)
		if err != nil {
			return fmt.Errorf("failed to add user for addUser: %w", err)
		}

		mailer := v.mailer.ConnectMailer()

		if mailer != nil {
			var tmpl *template.Template
			// TODO reimplement this
			// tmpl, err = v.template.GetEmailTemplate(templates.SignupEmailTemplate)
			if err != nil {
				return fmt.Errorf("failed to get email in addUser: %w", err)
			}

			file := mail.Mail{
				Subject: "Welcome to Aldermaston AFC!",
				Tpl:     tmpl,
				To:      u.Email,
				From:    "Aldermaston AFC No-Reply <no-reply.afc@bswdi.co.uk>",
				TplData: struct {
					Name     string
					Email    string
					Password string
				}{
					Name:     name,
					Email:    email,
					Password: password,
				},
			}

				err = mailer.SendMail(file)
				if err != nil {
					c1.Message = html.UnescapeString(fmt.Sprintf("successfully created user - failed to send email. Please send the username and password to this email: %s, password: %s", email, password))
					c1.MsgType = "is-warning"
					log.Printf("failed to send email for userAdd: %+v", err)
					log.Println("proceeding")
				} else {
					c1.Message = fmt.Sprintf("successfully created user, sent signup email to: \"%s\"", email)
					c1.MsgType = "is-success"
				}
			}
		} else {
			c1.Message = html.UnescapeString(fmt.Sprintf("successfully created user - failed to send email. Please send the username and password to this email: %s, password: %s", email, password))
			c1.MsgType = "is-warning"
			log.Println("no mailer present")
			log.Println("proceeding")
		}
		log.Printf("created user: %s, by: %d - %s", u.Email, c1.User.ID, c1.User.Email)

		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for uploadImage: %+v", err)
		}

		return c.JSON(status, message)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) UserEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse id for userEdit: %w", err)
		}
		userDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for userEdit: %w", err)
		}

		verifier := emailverifier.NewVerifier()

		var data struct {
			Error string `json:"error"`
		}

		tempName := c.FormValue("name")
		tempEmail := c.FormValue("email")
		tempPhone := c.FormValue("phone")

		userDB.Phone = null.NewString(tempPhone, len(tempPhone) > 0)

		if len(tempName) == 0 {
			log.Println("name must not be empty")
			data.Error = fmt.Sprintf("name must not be empty")
			return c.JSON(http.StatusOK, data)
		}

		userDB.Name = tempName

		res, err := verifier.Verify(tempEmail)
		if err != nil {
			log.Printf("failed to parse email for userEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to parse email for userEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		if !res.Syntax.Valid {
			log.Println("failed to parse email for userEdit: syntax is invalid")
			data.Error = fmt.Sprintf("failed to parse email for userEdit: syntax is invalid")
			return c.JSON(http.StatusOK, data)
		}

		userDB.Email = tempEmail

		formRole, err := role.GetRole(c.FormValue("role"))
		if err != nil {
			log.Printf("failed to get role for userEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to get role for userEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		userDB.Role = formRole

		teamID, err := strconv.Atoi(c.FormValue("userTeam"))
		if err != nil {
			log.Printf("failed to get teamID for userEdit: %+v, proceeding with no team", err)
			teamID = 0
		}
		if teamID < 0 {
			log.Println("failed to parse negative number, proceeding with no team")
			teamID = 0
		}

		if formRole.String() == role.Manager.String() {
			_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for userEdit: %+v, id: %d", err, teamID)
				data.Error = fmt.Sprintf("failed to get team for userEdit: %+v, id: %d", err, teamID)
				return c.JSON(http.StatusOK, data)
			}
		} else {
			teamID = 0
		}

		userDB.TeamID = teamID

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for userEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for userEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for userEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for userEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for userEdit: %+v", err)
				}
			}
			userDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveUserImage := c.FormValue("removeUserImage")
		if tempRemoveUserImage == "Y" {
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for userEdit: %+v", err)
				}
			}
			userDB.FileName = null.NewString("", false)
		} else if len(tempRemoveUserImage) != 0 {
			log.Printf("failed to parse removeUserImage for userEdit: %s", tempRemoveUserImage)
			data.Error = fmt.Sprintf("failed to parse removeUserImage for userEdit: %s", tempRemoveUserImage)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.user.EditUser(c.Request().Context(), userDB)
		if err != nil {
			log.Printf("failed to edit user for userEdit: %+v", err)
			data.Error = fmt.Sprintf("failed to edit user for userEdit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", userDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for userEdit: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for userDelete: %w", err)
		}

		userDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get user for userDelete: %w", err)
		}

		if userDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete user image for userDelete: %+v", err)
			}
		}

		err = v.user.DeleteUser(c.Request().Context(), userDB)
		if err != nil {
			return fmt.Errorf("failed to delete user for userDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", userDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for userDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/users")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
