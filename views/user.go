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
	"github.com/COMTOP1/AFC-GO/setting"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (v *Views) UsersFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	usersDB, err := v.user.GetUsers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get teams for users: %w", err)
	}

	displayEmail, err := v.setting.GetSetting(c.Request().Context(), "displayEmail")
	if err != nil {
		log.Printf("failed to get displayEmail for users, error: %+v, continuing", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		DisplayEmail string
		Users        []UserTemplate
		Teams        []TeamTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		DisplayEmail: displayEmail.SettingText,
		Users:        DBUsersToTemplateFormat(usersDB),
		Teams:        DBTeamsToTemplateFormat(teamsDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.UsersTemplate, templates.RegularType)
}

func (v *Views) UsersSetDisplayEmailFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		verifier := emailverifier.NewVerifier()

		var data struct {
			Error string `json:"error"`
		}

		tempEmail := c.FormValue("email")

		if len(tempEmail) != 0 {
			res, err := verifier.Verify(tempEmail)
			if err != nil {
				log.Printf("failed to parse email for display email, error: %+v", err)
				data.Error = fmt.Sprintf("failed to parse email for display email: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if !res.Syntax.Valid {
				log.Println("failed to parse email for display email: syntax is invalid")
				data.Error = "failed to parse email for display email: syntax is invalid"
				return c.JSON(http.StatusOK, data)
			}

			_, err = v.setting.EditSetting(c.Request().Context(), setting.Setting{
				ID:          "displayEmail",
				SettingText: tempEmail,
			})
			if err != nil {
				log.Printf("failed to edit setting for display email, error: %+v", err)
				data.Error = fmt.Sprintf("failed to edit setting for display email: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		} else {
			err := v.setting.DeleteSetting(c.Request().Context(), "displayEmail")
			if err != nil {
				log.Printf("failed to delete setting for display email, error: %+v", err)
			}
		}

		c1.Message = "successfully edited display email"
		c1.MsgType = "is-success"
		err := v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for display email, error: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
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

		if len(name) == 0 {
			log.Println("name must not be empty")
			data.Error = "name must not be empty"
			return c.JSON(http.StatusOK, data)
		}

		res, err := verifier.Verify(email)
		if err != nil {
			log.Printf("failed to parse email for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse email for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		if !res.Syntax.Valid {
			log.Println("failed to parse email for user add, error: syntax is invalid")
			data.Error = "failed to parse email for user add: syntax is invalid"
			return c.JSON(http.StatusOK, data)
		}

		formRole, err := role.GetRole(c.FormValue("role"))
		if err != nil {
			log.Printf("failed to get role for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to get role for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		teamID, err := strconv.Atoi(c.FormValue("userTeam"))
		if err != nil {
			log.Printf("failed to get teamID for user add, error: %+v, proceeding with no team", err)
			teamID = 0
		}
		if teamID < 0 {
			log.Println("failed to parse negative number, proceeding with no team")
			teamID = 0
		}

		if formRole.String() == role.Manager.String() {
			_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for user add, team id: %d, error: %+v", teamID, err)
				data.Error = fmt.Sprintf("failed to get team for user add, team id: %d: %+v", teamID, err)
				return c.JSON(http.StatusOK, data)
			}
		} else {
			teamID = 0
		}

		password, err := utils.GenerateRandom(utils.GeneratePassword)
		if err != nil {
			log.Printf("failed to generate password for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to generate password for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		salt, err := utils.GenerateRandom(utils.GenerateSalt)
		if err != nil {
			log.Printf("failed to generate salt for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to generate salt for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), v.conf.Security.ScryptWorkFactor, v.conf.Security.ScryptBlockSize, v.conf.Security.ScryptParallelismFactor, v.conf.Security.KeyLength)
		if err != nil {
			log.Printf("failed to generate password hash for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to generate password hash for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for user add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for user add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for user add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for user add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		u := user.User{
			Name:          name,
			Email:         email,
			Phone:         null.NewString(phone, len(phone) > 0),
			TeamID:        teamID,
			Role:          formRole,
			FileName:      null.NewString(fileName, hasUpload),
			ResetPassword: true,
			Hash:          null.StringFrom(hash),
			Salt:          null.NewString(salt, len(salt) > 0),
		}

		_, err = v.user.AddUser(c.Request().Context(), u)
		if err != nil {
			log.Printf("failed to add user for user add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to add user for user add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		mailer := v.mailer.ConnectMailer()

		if mailer != nil {
			var tmpl *template.Template
			tmpl, err = v.template.GetEmailTemplate(templates.SignupEmailTemplate)
			if err != nil {
				c1.Message = html.UnescapeString(fmt.Sprintf("successfully created user - no mailer present. Please send the username and password to this email: %s, password: %s", email, password))
				c1.MsgType = "is-warning"
				log.Printf("failed to get email for user add, error: %+v", err)
				log.Println("proceeding")
			} else {

				mailFile := mail.Mail{
					Subject: "Welcome to AFC Aldermaston!",
					Tpl:     tmpl,
					To:      u.Email,
					From:    "Aldermaston AFC No-Reply <no-reply.afc@bswdi.co.uk>",
					TplData: struct {
						Name     string
						Email    string
						Password string
						Domain   string
					}{
						Name:     name,
						Email:    email,
						Password: password,
						Domain:   v.conf.DomainName,
					},
				}

				err = mailer.SendMail(mailFile)
				if err != nil {
					c1.Message = html.UnescapeString(fmt.Sprintf("successfully created user - failed to send email. Please send the username and password to this email: %s, password: %s", email, password))
					c1.MsgType = "is-warning"
					log.Printf("failed to send email for user add, error: %+v", err)
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

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) UserEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to parse id for user edit, error: %w", err)
		}

		userDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for user edit, user id: %d, error: %w", userID, err)
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
			data.Error = "name must not be empty"
			return c.JSON(http.StatusOK, data)
		}

		userDB.Name = tempName

		res, err := verifier.Verify(tempEmail)
		if err != nil {
			log.Printf("failed to parse email for user edit, user id: %d, error: %+v", userID, err)
			data.Error = fmt.Sprintf("failed to parse email for user edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		if !res.Syntax.Valid {
			log.Println("failed to parse email for user edit: syntax is invalid")
			data.Error = "failed to parse email for user edit: syntax is invalid"
			return c.JSON(http.StatusOK, data)
		}

		userDB.Email = tempEmail

		formRole, err := role.GetRole(c.FormValue("role"))
		if err != nil {
			log.Printf("failed to get role for user edit, user id: %d, error: %+v", userID, err)
			data.Error = fmt.Sprintf("failed to get role for user edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		userDB.Role = formRole

		teamID, err := strconv.Atoi(c.FormValue("userTeam"))
		if err != nil {
			log.Printf("failed to get teamID for user edit, user id: %d, error: %+v, proceeding with no team", userID, err)
			teamID = 0
		}
		if teamID < 0 {
			log.Println("failed to parse negative number, proceeding with no team")
			teamID = 0
		}

		if formRole.String() == role.Manager.String() {
			_, err = v.team.GetTeam(c.Request().Context(), team.Team{ID: teamID})
			if err != nil {
				log.Printf("failed to get team for user edit, user id: %d, team id: %d, error: %+v", userID, teamID, err)
				data.Error = fmt.Sprintf("failed to get team for user edit, team id: %d: %+v", teamID, err)
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
				log.Printf("failed to get file for user edit, user id: %d, error: %+v", userID, err)
				data.Error = fmt.Sprintf("failed to get file for user edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for user edit, user id: %d, error: %+v", userID, err)
				data.Error = fmt.Sprintf("failed to upload file for user edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
			userDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveUserImage := c.FormValue("removeUserImage")
		if tempRemoveUserImage == "Y" {
			if userDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete image for user edit, user id: %d, error: %+v", userID, err)
				}
			}
			userDB.FileName = null.NewString("", false)
		} else if len(tempRemoveUserImage) != 0 {
			log.Printf("failed to parse removeUserImage for user edit, user id: %d, error: %s", userID, tempRemoveUserImage)
			data.Error = "failed to parse removeUserImage for user edit: " + tempRemoveUserImage
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.user.EditUser(c.Request().Context(), userDB)
		if err != nil {
			log.Printf("failed to edit user for user edit, user id: %d, error: %+v", userID, err)
			data.Error = fmt.Sprintf("failed to edit user for user edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", userDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for user edit, user id: %d, error: %+v", userID, err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for user delete, error: %w", err)
		}

		userDB, err := v.user.GetUser(c.Request().Context(), user.User{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get user for user delete, user id: %d, error: %w", id, err)
		}

		if userDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, userDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete user image for user delete, user id: %d, error: %+v", id, err)
			}
		}

		err = v.user.DeleteUser(c.Request().Context(), userDB)
		if err != nil {
			return fmt.Errorf("failed to delete user for user delete, user id: %d, error: %w", id, err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", userDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for user delete, user id: %d, error: %+v", id, err)
		}

		return c.Redirect(http.StatusFound, "/users")
	}
	return v.invalidMethodUsed(c)
}
