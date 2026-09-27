package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/document"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) DocumentsFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DocumentsFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	var d1 []document.Document
	var err error

	d1, err = v.document.GetDocuments(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get documents",
			fmt.Errorf("failed to get documents for documents: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Documents    []DocumentTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Documents:    DBDocumentsToTemplateFormat(d1),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.DocumentsTemplate, templates.RegularType)
}

func (v *Views) DocumentAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DocumentAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	file, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for document add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.documentSvc.Create(c.Request().Context(), document.CreateInput{Name: c.FormValue("name")}, file)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add document for document add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add document for document add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) DocumentDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.DocumentDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for document delete, error: %w", err)
	}
	deleted, err := v.documentSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete document for document delete, document id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/documents")
}
