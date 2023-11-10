package views

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

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

func DBNewsToTemplateFormat(n1 []news.News) []NewsTemplate {
	newsTemplate := make([]NewsTemplate, 0, len(n1))
	for _, newsDB := range n1 {
		var n2 NewsTemplate
		n2.ID = newsDB.ID
		n2.Title = newsDB.Title
		n2.Date = time.UnixMilli(newsDB.Temp).Format("2006-01-02 15:04:05 MST")
		newsTemplate = append(newsTemplate, n2)
	}
	return newsTemplate
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

func DBNewsToArticleTemplateFormat(newsDB news.News) NewsTemplate {
	var n NewsTemplate
	n.ID = newsDB.ID
	n.Title = newsDB.Title
	n.Content = newsDB.Content.String
	n.Date = time.UnixMilli(newsDB.Temp).Format("2006-01-02 15:04:05 MST")
	return n
}

func (v *Views) NewsAddFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) NewsEditFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}

func (v *Views) NewsDeleteFunc(c echo.Context) error {
	_ = c
	return fmt.Errorf("not implemented yet")
}
