package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) SponsorsFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.SponsorsFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	sponsorsDB, err := v.sponsor.GetSponsors(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get sponsors",
			fmt.Errorf("failed to get sponsors for sponsors, error: %w", err))
	}

	teamsDB, err := v.team.GetTeams(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get teams in sponsors",
			fmt.Errorf("failed to get teams for sponsors, error: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Sponsors     []SponsorTemplate
		Teams        []TeamTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Sponsors:     DBSponsorsToTemplateFormat(sponsorsDB),
		Teams:        DBTeamsToTemplateFormat(teamsDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.SponsorsTemplate, templates.RegularType)
}

func (v *Views) SponsorAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.SponsorAddFunc")
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
		data.Error = fmt.Sprintf("failed to get file for sponsor add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	created, err := v.sponsorSvc.Create(c.Request().Context(), sponsor.CreateInput{
		Name:    c.FormValue("name"),
		Website: c.FormValue("website"),
		Purpose: c.FormValue("purpose"),
		Team:    c.FormValue("team"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add sponsor for sponsor add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add sponsor for sponsor add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", created.Name))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) SponsorDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.SponsorDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for sponsor delete, error: %w", err)
	}
	deleted, err := v.sponsorSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete sponsor for sponsor delete, sponsor id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", deleted.Name))
	return c.Redirect(http.StatusFound, "/sponsors")
}
