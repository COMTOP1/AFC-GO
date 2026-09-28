package views

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
)

func (v *Views) AccountFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.AccountFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		User         UserTemplate
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		User:         DBUserToTemplateFormat(c1.User),
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.AccountTemplate, templates.RegularType)
}

func (v *Views) UploadImageFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.UploadImageFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		image, err := legacyUpload(c, "upload")
		if err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}
		if image == nil {
			data.Error = "failed to get file for upload image"
			return c.JSON(http.StatusOK, data)
		}

		if _, err = v.accountSvc.SetImage(c.Request().Context(), c1.User, image); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}

		v.flash(c, c1, "successfully uploaded image")
		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) RemoveImageFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.RemoveImageFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{}

		if _, err := v.accountSvc.RemoveImage(c.Request().Context(), c1.User); err != nil {
			data.Error = err.Error()
			return c.JSON(http.StatusOK, data)
		}

		v.flash(c, c1, "successfully removed image")
		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}
