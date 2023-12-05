package views

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
	"github.com/COMTOP1/AFC-GO/whatson"
)

func (v *Views) WhatsOnFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var w1 []whatson.WhatsOn
	var err error

	w1, err = v.whatsOn.GetWhatsOn(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get whatsOn for whatsOn: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		WhatsOn []WhatsOnTemplate
		User    user.User
		Context *Context
	}{
		Year:    year,
		WhatsOn: DBWhatsOnToTemplateFormat(w1),
		User:    c1.User,
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func (v *Views) WhatsOnArticleFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whatsOnArticle: %w", err))
	}
	whatsOnFromDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
	if err != nil {
		return fmt.Errorf("failed to get whatsOn for whatsOnArticle: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year    int
		WhatsOn WhatsOnTemplate
		User    user.User
		Context *Context
	}{
		Year:    year,
		WhatsOn: DBWhatsOnToArticleTemplateFormat(whatsOnFromDB),
		User:    c1.User,
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnArticleTemplate, templates.RegularType)
}

func (v *Views) WhatsOnAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		title := c.FormValue("title")
		content := c.FormValue("content")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		dateOfEvent := c.Request().FormValue("dateOfEvent")

		dateOfEventParsed, err := time.Parse("02/01/2006", dateOfEvent)
		if err != nil {
			log.Printf("failed to parse dateOfEvent for whatsOnAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfEvent for whatsOnAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for whatsOnAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for whatsOnAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whatsOnAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for whatsOnAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.whatsOn.AddWhatsOn(c.Request().Context(), whatson.WhatsOn{Title: title, Content: null.NewString(content, len(content) > 0), FileName: null.NewString(fileName, len(fileName) > 0), DateOfEvent: dateOfEventParsed})
		if err != nil {
			log.Printf("failed to add whatsOn for whatsOnAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add whatsOn for whatsOnAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whatsOnAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) WhatsOnEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		whatsOnID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whatsOnEdit: %w", err))
		}
		whatsOnDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
		if err != nil {
			return fmt.Errorf("failed to get whatsOn for whatsOnEdit: %w", err)
		}

		whatsOnDB.Title = c.FormValue("title")
		tempContent := c.FormValue("content")
		whatsOnDB.Content = null.NewString(tempContent, len(tempContent) > 0)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		dateOfEvent := c.Request().FormValue("dateOfEvent")

		tempDateOfEventParsed, err := time.Parse("02/01/2006", dateOfEvent)
		if err != nil {
			log.Printf("failed to parse dateOfEvent for whatsOnAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfEvent for whatsOnAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		whatsOnDB.DateOfEvent = tempDateOfEventParsed

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for whatsOnEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for whatsOnEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whatsOnEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for whatsOnEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if whatsOnDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for whatsOnEdit: %+v", err)
				}
			}
			whatsOnDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveWhatsOnImage := c.FormValue("removeWhatsOnImage")
		if tempRemoveWhatsOnImage == "Y" {

			whatsOnDB.FileName = null.NewString("", false)
		} else if len(tempRemoveWhatsOnImage) != 0 {
			log.Printf("failed to parse removeWhatsOnImage for whatsOnEdit: %s", tempRemoveWhatsOnImage)
			data.Error = fmt.Sprintf("failed to parse removeWhatsOnImage for whatsOnEdit: %s", tempRemoveWhatsOnImage)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.whatsOn.EditWhatsOn(c.Request().Context(), whatsOnDB)
		if err != nil {
			log.Printf("failed to add whatsOn for whatsOnAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add whatsOn for whatsOnAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", whatsOnDB.Title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whatsOnEdit: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) WhatsOnDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for whatsOnDelete: %w", err)
		}

		whatsOnDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get whatsOn for whatsOnDelete: %w", err)
		}

		if whatsOnDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete whatsOn image for whatsOnDelete: %+v", err)
			}
		}

		err = v.whatsOn.DeleteWhatsOn(c.Request().Context(), whatsOnDB)
		if err != nil {
			return fmt.Errorf("failed to delete whatsOn for whatsOnDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", whatsOnDB.Title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whatsOnDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/whatson")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
