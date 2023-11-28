package views

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/news"
	"github.com/COMTOP1/AFC-GO/templates"
	"github.com/COMTOP1/AFC-GO/user"
)

func (v *Views) NewsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	var n1 []news.News
	var err error

	n1, err = v.news.GetNews(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get news for news: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year int
		News []NewsTemplate
		User user.User
	}{
		Year: year,
		News: DBNewsToTemplateFormat(n1),
		User: c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.NewsTemplate, templates.RegularType)
}

func (v *Views) NewsArticleFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for newsArticle: %w", err))
	}
	newsDB, err := v.news.GetNewsArticle(c.Request().Context(), news.News{ID: newsID})
	if err != nil {
		return fmt.Errorf("failed to get news for newsArticle: %w", err)
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year int
		News NewsTemplate
		User user.User
	}{
		Year: year,
		News: DBNewsToArticleTemplateFormat(newsDB),
		User: c1.User,
	}

	return v.template.RenderTemplate(c.Response().Writer, data, templates.NewsArticleTemplate, templates.RegularType)
}

func (v *Views) NewsAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		title := c.FormValue("title")
		content := c.FormValue("content")

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		var fileName string
		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for newsAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for newsAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			fileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for newsAdd: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for newsAdd: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		_, err = v.news.AddNews(c.Request().Context(), news.News{Title: title, Content: null.NewString(content, len(content) > 0), FileName: null.StringFrom(fileName)})
		if err != nil {
			log.Printf("failed to add news for newsAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add news for newsAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) NewsEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		newsID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for newsEdit: %w", err))
		}
		newsDB, err := v.news.GetNewsArticle(c.Request().Context(), news.News{ID: newsID})
		if err != nil {
			return fmt.Errorf("failed to get news for newsEdit: %w", err)
		}

		newsDB.Title = c.FormValue("title")
		tempContent := c.FormValue("content")
		newsDB.Content = null.NewString(tempContent, len(tempContent) > 0)

		data := struct {
			Error string `json:"error"`
		}{
			Error: "",
		}

		hasUpload := true

		file, err := c.FormFile("upload")
		if err != nil {
			if !strings.Contains(err.Error(), "no such file") {
				log.Printf("failed to get file for newsEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to get file for newsEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			hasUpload = false
		}
		if hasUpload {
			var tempFileName string
			tempFileName, err = v.fileUpload(file)
			if err != nil {
				log.Printf("failed to upload file for newsEdit: %+v", err)
				data.Error = fmt.Sprintf("failed to upload file for newsEdit: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
			newsDB.FileName = null.StringFrom(tempFileName)
		}

		_, err = v.news.EditNews(c.Request().Context(), newsDB)
		if err != nil {
			log.Printf("failed to add news for newsAdd: %+v", err)
			data.Error = fmt.Sprintf("failed to add news for newsAdd: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		return c.JSON(http.StatusOK, data)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

func (v *Views) NewsDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return fmt.Errorf("failed to get id for newsDelete: %w", err)
		}

		newsDB, err := v.news.GetNewsArticle(c.Request().Context(), news.News{ID: id})
		if err != nil {
			return fmt.Errorf("failed to get news for newsDelete: %w", err)
		}

		if newsDB.FileName.Valid {
			err = os.Remove(filepath.Join(v.conf.FileDir, newsDB.FileName.String))
			if err != nil {
				return fmt.Errorf("failed to delete news image for newsDelete: %w", err)
			}
		}

		err = v.news.DeleteNews(c.Request().Context(), newsDB)
		if err != nil {
			return fmt.Errorf("failed to delete news for newsDelete: %w", err)
		}
		return c.Redirect(http.StatusFound, "/news")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
