package views

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
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

func (v *Views) invalidMethodUsed(c echo.Context) *echo.HTTPError {
	return &echo.HTTPError{
		Code:     http.StatusMethodNotAllowed,
		Message:  "invalid method used",
		Internal: fmt.Errorf("invalid method used, path: %s, method: %s", c.Path(), c.Request().Method),
	}
}

func (v *Views) error(code int, publicMessage string, internalMessage error) *echo.HTTPError {
	return &echo.HTTPError{
		Code:     code,
		Message:  publicMessage,
		Internal: internalMessage,
	}
}
