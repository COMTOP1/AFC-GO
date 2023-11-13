package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/image"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) GalleryFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	imagesDB, err := v.image.GetImages(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get images for gallery: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year   int
		Images []image.Image
		User   user.User
	}{
		Year:   year,
		Images: imagesDB,
		User:   c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.GalleryTemplate, templates.RegularType)
}

func (v *Views) ImageAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) ImageDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
