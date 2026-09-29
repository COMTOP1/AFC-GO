package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/affiliation"
)

func (v *Views) AffiliationAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.AffiliationAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for affiliation add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.affiliationSvc.Create(c.Request().Context(), affiliation.CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add affiliation for affiliation add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add affiliation for affiliation add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) AffiliationDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.AffiliationDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for affiliation delete, error: %w", err)
	}
	deleted, err := v.affiliationSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete affiliation for affiliation delete, affiliation id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/")
}
