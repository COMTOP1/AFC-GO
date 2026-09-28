package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/image"
	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) GalleryFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.GalleryFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	imagesDB, err := v.image.GetImages(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get gallery images",
			fmt.Errorf("failed to get images for gallery, error: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		Images       []image.Image
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		Images:       imagesDB,
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.GalleryTemplate, templates.RegularType)
}

func (v *Views) ImageAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ImageAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}
	photo, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for image add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	if _, err = v.gallerySvc.Create(c.Request().Context(), image.CreateInput{Caption: c.FormValue("caption")}, photo); err != nil {
		slog.Info(fmt.Sprintf("failed to add image for image add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add image for image add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	v.flash(c, c1, "successfully added image")
	return c.JSON(http.StatusOK, data)
}

func (v *Views) ImageDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.ImageDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for image delete, error: %w", err)
	}
	if _, err = v.gallerySvc.Delete(c.Request().Context(), id); err != nil {
		return fmt.Errorf("failed to delete image for image delete, image id: %d, error: %w", id, err)
	}
	v.flash(c, c1, "successfully deleted image")
	return c.Redirect(http.StatusFound, "/gallery")
}
