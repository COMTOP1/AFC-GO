package views

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
)

func (v *Views) AccountFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		User    UserTemplate
		Context *Context
	}{
		Year:    year,
		User:    DBUserToTemplateFormat(c1.User),
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.AccountTemplate, templates.RegularType)
}

func (v *Views) UploadImageFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		if c1.User.FileName.Valid {
			err := os.Remove(filepath.Join(v.conf.FileDir, c1.User.FileName.String))
			if err != nil {
				log.Printf("failed to delete image for uploadImage: %+v", err)
			}
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for uploadImage: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for uploadImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		var fileName string
		fileName, err = v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for uploadImage: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for uploadImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.User.FileName = null.NewString(fileName, len(fileName) > 0)

		err = v.user.EditUserImage(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to edit user for uploadImage: %+v", err)
			data.Error = fmt.Sprintf("failed to edit user for uploadImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully uploaded image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for uploadImage: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) RemoveImageFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		if c1.User.FileName.Valid {
			err := os.Remove(filepath.Join(v.conf.FileDir, c1.User.FileName.String))
			if err != nil {
				log.Printf("failed to delete image for removeImage: %+v", err)
			}
		}

		c1.User.FileName = null.NewString("", false)

		err := v.user.EditUserImage(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to edit user for removeImage: %+v", err)
			data.Error = fmt.Sprintf("failed to edit user for removeImage: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully removed image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for removedImage: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
