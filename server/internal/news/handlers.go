package news

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/COMTOP1/AFC-GO/server/internal/web"
)

// Handlers serves /news.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// Register mounts the news routes on the /api/v1 group.
func (h *Handlers) Register(g *echo.Group, guards web.Guards) {
	g.GET("/news", h.list)
	g.GET("/news/:id", h.get)
	g.POST("/news", h.create, guards.Editor)
	g.PATCH("/news/:id", h.update, guards.Editor)
	g.DELETE("/news/:id", h.remove, guards.Editor)
}

// list returns every article, newest first.
//
//	@Summary	List news
//	@Tags		news
//	@Produce	json
//	@Success	200	{array}	news.Article
//	@Router		/news [get]
func (h *Handlers) list(c echo.Context) error {
	out, err := h.svc.List(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
}

// get returns one article.
//
//	@Summary	Get a news article
//	@Tags		news
//	@Produce	json
//	@Param		id	path		int	true	"Article ID"
//	@Success	200	{object}	news.Article
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/news/{id} [get]
func (h *Handlers) get(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	a, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, a)
}

// create adds an article.
//
//	@Summary	Create a news article
//	@Tags		news
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		title	formData	string	true	"Title"
//	@Param		content	formData	string	false	"HTML content (sanitised)"
//	@Param		image	formData	file	false	"Image"
//	@Success	201		{object}	news.Article
//	@Failure	422		{object}	web.ErrorResponse
//	@Router		/news [post]
func (h *Handlers) create(c echo.Context) error {
	if err := web.RequireForm(c); err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	a, err := h.svc.Create(c.Request().Context(), CreateInput{
		Title:   c.FormValue("title"),
		Content: c.FormValue("content"),
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, a)
}

// update changes an article; omitted fields are left as they are.
//
//	@Summary	Update a news article
//	@Tags		news
//	@Accept		multipart/form-data
//	@Produce	json
//	@Param		id			path		int		true	"Article ID"
//	@Param		title		formData	string	false	"Title"
//	@Param		content		formData	string	false	"HTML content (sanitised); empty clears it"
//	@Param		image		formData	file	false	"Replacement image"
//	@Param		removeImage	formData	bool	false	"Remove the current image"
//	@Success	200			{object}	news.Article
//	@Failure	404			{object}	web.ErrorResponse
//	@Failure	422			{object}	web.ErrorResponse
//	@Router		/news/{id} [patch]
func (h *Handlers) update(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if err = web.RequireForm(c); err != nil {
		return err
	}
	image, err := web.FormFile(c, "image")
	if err != nil {
		return err
	}
	remove, err := web.FormBool(c, "removeImage")
	if err != nil {
		return err
	}
	a, err := h.svc.Update(c.Request().Context(), id, UpdateInput{
		Title:       web.FormString(c, "title"),
		Content:     web.FormString(c, "content"),
		RemoveImage: remove != nil && *remove,
	}, image)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, a)
}

// remove deletes an article.
//
//	@Summary	Delete a news article
//	@Tags		news
//	@Param		id	path	int	true	"Article ID"
//	@Success	204
//	@Failure	404	{object}	web.ErrorResponse
//	@Router		/news/{id} [delete]
func (h *Handlers) remove(c echo.Context) error {
	id, err := web.ParamID(c, "id")
	if err != nil {
		return err
	}
	if _, err = h.svc.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
