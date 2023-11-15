package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/COMTOP1/AFC-GO/document"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

type DocumentTemplate struct {
	ID   int
	Name string
}

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
	}{
		Year:      year,
		Documents: DBDocumentsToTemplateFormat(d1),
		User:      c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.DocumentsTemplate, templates.RegularType)
}

func (v *Views) DocumentAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
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
			// return fmt.Errorf("failed to upload file for documentAdd: %w", err)
			log.Printf("failed to upload file for documentAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		_, err = v.document.AddDocument(c.Request().Context(), document.Document{Name: name, FileName: fileName})
		if err != nil {
			log.Printf("failed to add document for documentAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add document for documentAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
			// return fmt.Errorf("failed to add document for documentAdd: %w", err)
		}

		return c.JSON(http.StatusOK, data)
		// return c.Redirect(http.StatusFound, "/")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) DocumentDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for documentDelete: %w", err)
		}

		document1, err := v.document.GetDocument(c.Request().Context(), document.Document{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get user for documentDelete: %w", err)
		}

		err = v.document.DeleteDocument(c.Request().Context(), document1)
		if err != nil {
			return fmt.Errorf("failed to delete user for documentDelete: %w", err)
		}
		return c.Redirect(http.StatusFound, "/documents")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
