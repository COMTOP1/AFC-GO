package views

import (
	"errors"
	"github.com/COMTOP1/AFC-GO/user"
	"log"
	"time"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/labstack/echo/v4"
)

func (v *Views) CustomHTTPErrorHandler(err error, c echo.Context) {
	c1 := v.getSessionData(c)
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
	year, _, _ := time.Now().Date()
	c.Response().WriteHeader(status)
	data := struct {
		Code  int
		Error any
		Year  int
		User  user.User
	}{
		Code:  status,
		Error: message,
		Year:  year,
		User:  c1.User,
	}
	err1 := v.template.RenderTemplate(c.Response().Writer, data, templates.ErrorTemplate, templates.RegularType)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}

func (v *Views) Error404(c echo.Context) error {
	return v.template.RenderTemplate(c.Response().Writer, nil, templates.NotFound404Template, templates.RegularType)
}
