package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) DocumentsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var d1 []document.Document
	var err error

	d1, err = v.document.GetDocuments(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get documents for documents: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year      int
		Documents []DocumentTemplate
		User      user.User
		Context   *Context
	}{
		Year:      year,
		Documents: DBDocumentsToTemplateFormat(d1),
		User:      c1.User,
		Context:   c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.DocumentsTemplate, templates.RegularType)
}

func (v *Views) DocumentAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		name := c.FormValue("name")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for documentAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for documentAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.document.AddDocument(c.Request().Context(), document.Document{Name: name, FileName: fileName})
		if err != nil {
			log.Printf("failed to add document for documentAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add document for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = fmt.Sprintf("successfully added \"%s\"", name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for documentAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}

func (v *Views) DocumentDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for documentDelete: %w", err)
		}

		documentDB, err := v.document.GetDocument(c.Request().Context(), document.Document{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get user for documentDelete: %w", err)
		}

		err = os.Remove(filepath.Join(v.conf.FileDir, documentDB.FileName))
		if err != nil {
			log.Printf("failed to delete document file for documentDelete: %+v", err)
		}

		err = v.document.DeleteDocument(c.Request().Context(), documentDB)
		if err != nil {
			return fmt.Errorf("failed to delete user for documentDelete: %w", err)
		}

		c1.Message = fmt.Sprintf("successfully deleted \"%s\"", documentDB.Name)
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for documentDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/documents")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, errors.New("invalid method used"))
}
