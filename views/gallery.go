package views

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

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
		Year    int
		Images  []image.Image
		User    user.User
		Context *Context
	}{
		Year:    year,
		Images:  imagesDB,
		User:    c1.User,
		Context: c1,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.GalleryTemplate, templates.RegularType)
}

func (v *Views) ImageAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		file, err := c.FormFile("upload")
		if err != nil {
			log.Printf("failed to get file for imageAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to get file for imageAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}
		fileName, err := v.fileUpload(file)
		if err != nil {
			log.Printf("failed to upload file for imageAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to upload file for imageAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		caption := c.FormValue("caption")

		_, err = v.image.AddImage(c.Request().Context(), image.Image{FileName: fileName, Caption: null.NewString(caption, len(caption) > 0)})
		if err != nil {
			log.Printf("failed to add image for imageAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add image for imageAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully added image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for imageAdd: %+v", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) ImageDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for imageDelete: %w", err)
		}

		imageDB, err := v.image.GetImage(c.Request().Context(), image.Image{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get image for imageDelete: %w", err)
		}

		err = os.Remove(filepath.Join(v.conf.FileDir, imageDB.FileName))
		if err != nil {
			log.Printf("failed to delete image file for imageDelete: %+v", err)
		}

		err = v.image.DeleteImage(c.Request().Context(), imageDB)
		if err != nil {
			return fmt.Errorf("failed to delete image for imageDelete: %w", err)
		}

		c1.Message = "successfully deleted image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for imageDelete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/gallery")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
