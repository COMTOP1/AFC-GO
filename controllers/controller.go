package controllers

import (
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/utils"
	"github.com/COMTOP1/api/handler"
	"github.com/labstack/echo/v4"
	"net/http"
)

// ControllerInterface is the interface to which controllers adhere.
type ControllerInterface interface {
	Get()     //method = GET processing
	Post()    //method = POST processing
	Delete()  //method = DELETE processing
	Put()     //method = PUT handling
	Head()    //method = HEAD processing
	Patch()   //method = PATCH treatment
	Options() //method = OPTIONS processing
	Connect() //method = CONNECT processing
	Trace()   //method = TRACE processing
}

// Controller is the base type of controllers in the 2016site architecture.
type Controller struct {
	Access   *utils.Accesser
	Template *templates.Templater
	Session  *handler.Session
}

func GetController(accesser *utils.Accesser, session *handler.Session) Controller {
	return Controller{Access: accesser, Session: session}
}

// Get handles a HTTP GET request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Get(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Post handles a HTTP POST request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Post(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Delete handles a HTTP DELETE request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Delete(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Put handles a HTTP PUT request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Put(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Head handles a HTTP HEAD request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Head(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Patch handles a HTTP PATCH request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Patch(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Options handles a HTTP OPTIONS request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Options(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Connect handles a HTTP CONNECT request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Connect(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Trace handles a HTTP TRACE request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Trace(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}
