package routes

import (
	"fmt"
	"github.com/COMTOP1/AFC-GO/controllers"
	_ "github.com/COMTOP1/AFC-GO/controllers"
	"github.com/COMTOP1/AFC-GO/middleware"
	"github.com/COMTOP1/AFC-GO/myUtils"
	"github.com/COMTOP1/AFC-GO/structs"
	"github.com/COMTOP1/AFC-GO/utils"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"
	"net/http"
)

type (
	Router struct {
		config *structs.Config
		port   string
		repos  *controllers.Repos
		router *echo.Echo
		access *utils.Accesser
		mailer *utils.Mailer
	}
	NewRouter struct {
		Config *structs.Config
		Port   string
		//DomainName string
		Repos    *controllers.Repos
		Debug    bool
		Accesser *utils.Accesser
		Mailer   *utils.Mailer
	}
)

func New(conf *NewRouter) *Router {
	myUtil := myUtils.MyUtils{}

	conf.Config.PageContext.GetTime = myUtil.GetTime
	conf.Config.PageContext.GetTime1Day = myUtil.GetTime1Day
	conf.Config.PageContext.GetDay = myUtil.GetDay
	conf.Config.PageContext.GetYear = myUtil.GetYear

	r := &Router{
		config: conf.Config,
		router: echo.New(),
		repos:  conf.Repos,
		access: conf.Accesser,
		mailer: conf.Mailer,
	}
	r.router.HideBanner = true

	r.router.Debug = r.config.Server.Debug

	middleware.New(r.router, r.config.Server.DomainName)

	r.loadRoutes()

	return r
}

func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Server.Port))
	return fmt.Errorf("failed to start router on port %s", r.config.Server.Port)
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", func(c echo.Context) error {
		return c.JSON(http.StatusNotFound, utils.Error{Error: "Not found"})
	})

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.GET("/", r.repos.Home.Home)

	r.router.GET("/home", r.repos.Home.Home)

	r.router.GET("/teams", r.repos.Teams.Teams)

	r.router.GET("/team", r.repos.Teams.Team)

	r.router.GET("/download", r.repos.Download.Download)

	r.router.GET("/public/:file", r.repos.Public.Public)

	r.router.GET("/public/webfonts/Arial/:file", r.repos.Public.PublicFontArial)

	r.router.GET("/public/webfonts/Allerta/:file", r.repos.Public.PublicFontAllerta)

	//r.router.GET("/set", func(c echo.Context) error {
	//	// Initialize a new cookie containing the string "Hello world!" and some
	//	// non-default attributes.
	//	_, err := c.Cookie("exampleCookie")
	//	if err == nil {
	//		return c.JSON(http.StatusBadRequest, utils.Error{Error: "cookie already exists"})
	//	}
	//
	//	cookie := http.Cookie{
	//		Name:     "exampleCookie",
	//		Value:    "Hello world!",
	//		Path:     "/",
	//		MaxAge:   10,
	//		HttpOnly: true,
	//		//Secure:   true,
	//		SameSite: http.SameSiteLaxMode,
	//	}
	//
	//	// Use the http.SetCookie() function to send the cookie to the client.
	//	// Behind the scenes this adds a `Set-Cookie` header to the response
	//	// containing the necessary cookie data.
	//	c.SetCookie(&cookie)
	//
	//	//// Write a HTTP response as normal.
	//	//write, err := w.Write([]byte("cookie set!"))
	//	//if err != nil {
	//	//	fmt.Println("Error - " + err.Error())
	//	//}
	//	//fmt.Println(write)
	//	return c.JSON(http.StatusOK, "wrote cookie")
	//})
	//
	//r.router.GET("/get", func(c echo.Context) error {
	//	// Retrieve the cookie from the request using its name (which in our case is
	//	// "exampleCookie"). If no matching cookie is found, this will return a
	//	// http.ErrNoCookie error. We check for this, and return a 400 Bad Request
	//	// response to the client.
	//	cookie, err := c.Cookie("exampleCookie")
	//	if err != nil {
	//		switch {
	//		case errors.Is(err, http.ErrNoCookie):
	//			//http.Error(w, "cookie not found", http.StatusBadRequest)
	//			return c.JSON(http.StatusBadRequest, utils.Error{Error: "cookie not found"})
	//		default:
	//			log.Println(err)
	//			//http.Error(w, "server error", http.StatusInternalServerError)
	//			return c.JSON(http.StatusInternalServerError, utils.Error{Error: fmt.Sprintf("server error: %v", err)})
	//		}
	//	}
	//
	//	// Echo out the cookie value in the response body.
	//	//w.Write([]byte(cookie.Value))
	//	return c.JSON(http.StatusOK, cookie.Value)
	//})
	//
	//r.router.GET("/delete", func(c echo.Context) error {
	//	cookie, err := c.Cookie("exampleCookie")
	//	if err != nil {
	//		switch {
	//		case errors.Is(err, http.ErrNoCookie):
	//			//http.Error(w, "cookie not found", http.StatusBadRequest)
	//			return c.JSON(http.StatusBadRequest, utils.Error{Error: "cookie not found"})
	//		default:
	//			log.Println(err)
	//			//http.Error(w, "server error", http.StatusInternalServerError)
	//			return c.JSON(http.StatusInternalServerError, utils.Error{Error: fmt.Sprintf("server error: %v", err)})
	//		}
	//	}
	//	cookie.MaxAge = 0
	//	cookie.Expires = time.Now()
	//	cookie.Value = ""
	//	c.SetCookie(cookie)
	//	return c.JSON(http.StatusOK, "deleted")
	//})
}
