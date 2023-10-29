package main

import (
	"embed"
	"github.com/labstack/echo/v4"
	"net/http"
)

//go:embed assets/*
var embeddedFiles embed.FS

//go:embed holding.html
var holdingHTML embed.FS

func main() {
	e := echo.New()
	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "assets")))

	e.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", assetHandler)))
	e.GET("/*", echo.StaticFileHandler("holding.html", holdingHTML))
	e.Logger.Fatal(e.Start(":1323"))
}
