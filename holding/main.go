package main

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v5"
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
	err := e.Start(":1323")
	e.Logger.Error("server could not start or failed", "err", err)
}
