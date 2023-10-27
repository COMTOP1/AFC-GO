package views

import (
	"errors"
	"log"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/labstack/echo/v4"
)

func (v *Views) CustomHTTPErrorHandler(err error, c echo.Context) {
	log.Print(err)
	var he *echo.HTTPError
	var status int
	if errors.As(err, &he) {
		status = he.Code
	} else {
		status = 500
	}
	var message interface{}
	message = err
	if he != nil {
		message = he.Message
	}
	c.Response().WriteHeader(status)
	data := struct {
		Code  int
		Error any
	}{
		Code:  status,
		Error: message,
	}
	err1 := v.template.RenderTemplate(c.Response().Writer, data, templates.ErrorTemplate, templates.NoNavType)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderTemplate(c.Response().Writer, nil, templates.NotFound404Template, templates.NoNavType)
}
