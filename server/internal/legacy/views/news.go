package views

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/legacy/templates"
	"github.com/COMTOP1/AFC-GO/server/internal/news"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

func (v *Views) NewsFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	var n1 []news.News
	var err error

	n1, err = v.news.GetNews(c.Request().Context())
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get news",
			fmt.Errorf("failed to get news for news: %w", err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		News         []NewsTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		News:         DBNewsToTemplateFormat(n1),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.NewsTemplate, templates.RegularType)
}

func (v *Views) NewsArticleFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsArticleFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	c1 := v.getSessionData(c)

	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return v.error(http.StatusBadRequest, "invalid id provided for news article",
			fmt.Errorf("failed to parse id for news article: %w", err))
	}

	newsDB, err := v.news.GetNewsArticle(c.Request().Context(), news.News{ID: newsID})
	if err != nil {
		return v.error(http.StatusInternalServerError, "failed to get news article",
			fmt.Errorf("failed to get news for news article, news id: %d, error: %w", newsID, err))
	}

	year, _, _ := time.Now().Date()

	data := struct {
		Year         int
		VisitorCount int
		News         NewsTemplate
		User         user.User
		Context      *Context
	}{
		Year:         year,
		VisitorCount: v.GetVisitorCount(),
		News:         DBNewsToArticleTemplateFormat(newsDB),
		User:         c1.User,
		Context:      c1,
	}

	return v.template.RenderTemplate(c.Request().Context(), c.Response().Writer, data, templates.NewsArticleTemplate, templates.RegularType)
}

func (v *Views) NewsAddFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsAddFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	data := struct {
		Error string `json:"error"`
	}{}

	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for news add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	article, err := v.newsSvc.Create(c.Request().Context(), news.CreateInput{
		Title:   c.FormValue("title"),
		Content: c.FormValue("htmlContent"),
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to add news for news add, error: %+v", err))
		data.Error = fmt.Sprintf("failed to add news for news add: %+v", err)
		return c.JSON(http.StatusOK, data)
	}

	v.flash(c, c1, fmt.Sprintf("successfully added \"%s\"", article.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) NewsEditFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsEditFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	newsID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse id for news edit, error: %w", err))
	}
	data := struct {
		Error string `json:"error"`
	}{}

	remove := c.FormValue("removeNewsImage")
	if remove != "" && remove != "Y" {
		data.Error = "failed to parse removeNewsImage for news edit: " + remove
		return c.JSON(http.StatusOK, data)
	}
	image, err := legacyUpload(c, "upload")
	if err != nil {
		data.Error = fmt.Sprintf("failed to get file for news edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}
	title, content := c.FormValue("title"), c.FormValue("htmlContent")
	article, err := v.newsSvc.Update(c.Request().Context(), newsID, news.UpdateInput{
		Title:       &title,
		Content:     &content,
		RemoveImage: remove == "Y",
	}, image)
	if err != nil {
		slog.Info(fmt.Sprintf("failed to edit news for news edit, news id: %d, error: %+v", newsID, err))
		data.Error = fmt.Sprintf("failed to edit news for news edit: %+v", err)
		return c.JSON(http.StatusOK, data)
	}

	v.flash(c, c1, fmt.Sprintf("successfully edited \"%s\"", article.Title))
	return c.JSON(http.StatusOK, data)
}

func (v *Views) NewsDeleteFunc(c echo.Context) error {
	spanCtx, span := tracer.Start(c.Request().Context(), "views.NewsDeleteFunc")
	defer span.End()
	c.SetRequest(c.Request().WithContext(spanCtx))
	if c.Request().Method != http.MethodPost {
		return v.invalidMethodUsed(c)
	}
	c1 := v.getSessionData(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get id for news delete, error: %w", err)
	}
	article, err := v.newsSvc.Delete(c.Request().Context(), id)
	if err != nil {
		return fmt.Errorf("failed to delete news for news delete, news id: %d, error: %w", id, err)
	}
	v.flash(c, c1, fmt.Sprintf("successfully deleted \"%s\"", article.Title))
	return c.Redirect(http.StatusFound, "/news")
}
