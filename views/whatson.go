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

	whatsOnsDB, err := v.whatsOn.GetWhatsOn(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get whats on articles",
			fmt.Errorf("failed to get whats ons for whatsOn, error: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year       int
		WhatsOn    []WhatsOnTemplate
		User       user.User
		TimePeriod string
		Selected   string
		Context    *Context
	}{
		Year:       year,
		WhatsOn:    DBWhatsOnToTemplateFormat(whatsOnsDB),
		User:       c1.User,
		TimePeriod: "all the what's on articles",
		Selected:   "all",
		Context:    c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func (v *Views) WhatsOnTomePeriodFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	timePeriod := c.Param("timePeriod")

	var whatsOnsDB []whatson.WhatsOn
	var err error

	switch timePeriod {
	case "future":
		whatsOnsDB, err = v.whatsOn.GetWhatsOnFuture(c.Request().Context())
		if err != nil {
			return v.error(http.StatusInternalServerError, "failed to get whats on articles in the future",
				fmt.Errorf("failed to get whats on future for whats on, error: %w", err))
		}
	case "past":
		whatsOnsDB, err = v.whatsOn.GetWhatsOnPast(c.Request().Context())
		if err != nil {
			return v.error(http.StatusInternalServerError, "failed to get whats on articles in the past",
				fmt.Errorf("failed to get whats on past for whats on, error: %w", err))
		}
	case "all":
		fallthrough
	default:
		return c.Redirect(http.StatusFound, "/whatson")
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year       int
		WhatsOn    []WhatsOnTemplate
		User       user.User
		TimePeriod string
		Selected   string
		Context    *Context
	}{
		Year:       year,
		WhatsOn:    DBWhatsOnToTemplateFormat(whatsOnsDB),
		User:       c1.User,
		TimePeriod: fmt.Sprintf("all the %s what's on articles", timePeriod),
		Selected:   timePeriod,
		Context:    c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.WhatsOnTemplate, templates.RegularType)
}

func (v *Views) WhatsOnSelectFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		timePeriod := c.FormValue("timePeriod")

		switch timePeriod {
		case "future":
			return c.Redirect(http.StatusFound, "/whatsonperiod/future")
		case "past":
			return c.Redirect(http.StatusFound, "/whatsonperiod/past")
		default:
			return c.Redirect(http.StatusFound, "/whatson")
		}
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) WhatsOnArticleFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	whatsOnID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return v.error(http.StatusBadRequest, "invalid id provided for whats on article",
			fmt.Errorf("failed to parse whats on id for whats on article, error: %w", err))
	}
	whatsOnFromDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
	if err != nil {
		return fmt.Errorf("failed to get whats on for whats on article, error: %w", err)
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
			log.Printf("failed to parse dateOfEvent for whats on add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for whats on add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for whats on add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whats on add, error: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for whats on add: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.whatsOn.AddWhatsOn(c.Request().Context(), whatson.WhatsOn{Title: title, Content: null.NewString(content, len(content) > 0), FileName: null.NewString(fileName, len(fileName) > 0), DateOfEvent: dateOfEventParsed})
		if err != nil {
			log.Printf("failed to add whatsOn for whats on add, error: %+v", err)
			data.Error = fmt.Sprintf("failed to add whatsOn for whats on add: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whats on add, error: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) WhatsOnEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		whatsOnID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for whats on edit, error: %w", err))
		}
		whatsOnDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: whatsOnID})
		if err != nil {
			return fmt.Errorf("failed to get whatsOn for whats on edit, whats on id: %d, error: %w", whatsOnID, err)
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
			log.Printf("failed to parse dateOfEvent for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
			data.Error = fmt.Sprintf("failed to parse dateOfEvent for whats on edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		whatsOnDB.DateOfEvent = tempDateOfEventParsed

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				data.Error = fmt.Sprintf("failed to get file for whats on edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				data.Error = fmt.Sprintf("failed to upload file for whats on edit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			if whatsOnDB.FileName.Valid {
				err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
				if err != nil {
					log.Printf("failed to delete old image for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
				}
			}
			whatsOnDB.FileName = null.NewString(tempFileName, len(tempFileName) > 0)
		}

		tempRemoveWhatsOnImage := c.FormValue("removeWhatsOnImage")
		if tempRemoveWhatsOnImage == "Y" {

			whatsOnDB.FileName = null.NewString("", false)
		} else if len(tempRemoveWhatsOnImage) != 0 {
			log.Printf("failed to parse removeWhatsOnImage for whats on edit, whats on id: %d, value: %s", whatsOnID, tempRemoveWhatsOnImage)
			data.Error = "failed to parse removeWhatsOnImage for whats on edit, value: " + tempRemoveWhatsOnImage
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.whatsOn.EditWhatsOn(c.Request().Context(), whatsOnDB)
		if err != nil {
			log.Printf("failed to add whatsOn for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
			data.Error = fmt.Sprintf("failed to add whatsOn for whats on edit: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully edited \"%s\"", whatsOnDB.Title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whats on edit, whats on id: %d, error: %+v", whatsOnID, err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) WhatsOnDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for whats on delete, error: %w", err)
		}

		whatsOnDB, err := v.whatsOn.GetWhatsOnArticle(c.Request().Context(), whatson.WhatsOn{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get whatsOn for whats on delete, whats on id: %d, error: %w", id, err)
		}

		if whatsOnDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, whatsOnDB.FileName.String))
			if err != nil {
				log.Printf("failed to delete whatsOn image for whats on delete, whats on id: %d, error: %+v", id, err)
			}
		}

		err = v.whatsOn.DeleteWhatsOn(c.Request().Context(), whatsOnDB)
		if err != nil {
			return fmt.Errorf("failed to delete whatsOn for whats on delete, whats on id: %d, error: %w", id, err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", whatsOnDB.Title)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for whats on delete, whats on id: %d, error: %+v", id, err)
		}

		return c.Redirect(http.StatusFound, "/whatson")
	}
	return v.invalidMethodUsed(c)
}
